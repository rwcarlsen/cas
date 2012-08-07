
package mount

import (
  "fmt"
  "os"
  "io/ioutil"
  "time"
  "encoding/json"
  "path/filepath"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobserv"
  "github.com/rwcarlsen/cas/query"
)

type Mount struct {
  Cl *blobserv.Client
  Root string
  BlobPath string
  Refs map[string]string
  UseHistory bool
  PathFor func(*blob.FileMeta)string `json:"-"`
  q *query.Query
}

func New(pathFn func(*blob.FileMeta)string, q *query.Query) *Mount {
  return &Mount{
    PathFor: pathFn,
    q: q,
  }
}

func (m *Mount) ConfigClient(user, pass, host string) {
  m.Cl = &blobserv.Client{
    User: user,
    Pass: pass,
    Host: host,
  }
}

func (m *Mount) Unpack() error {
  err := m.Cl.Dial()
  if err != nil {
    return err
  }
  m.runQuery()

  m.Refs = map[string]string{}
  for _, b := range m.q.Results {
    fm := &blob.FileMeta{}
    err := blob.Unmarshal(b, fm)
    if err != nil || (!m.UseHistory && !m.isTip(b)) {
      continue
    }

    fm, data, err := m.Cl.ReconstituteFile(b.Ref())
    if err != nil {
      return err
    }

    pth := filepath.Join(m.Root, m.PathFor(fm))
    os.MkdirAll(filepath.Dir(pth), 0744)
    f, err := os.Create(pth)
    if err != nil {
      return err
    }
    f.Write(data)
    f.Close()
    m.Refs[m.PathFor(fm)] = fm.RcasObjectRef
  }
  return nil
}

func (m *Mount) Snapshot() error {
  var bigerr error
  fn := func(path string, info os.FileInfo, inerr error) error {
    if info.IsDir() {
      return nil
    }

    if path[0] != '/' {
      path = "/" + path
    }

    fmt.Println("path=", path)
    var newfm = &blob.FileMeta{}
    var chunks []*blob.Blob
    if ref, ok := m.Refs[path]; ok {
      fmt.Println("preexisting")
      b, err := m.Cl.ObjectTip(ref)
      if err != nil {
        bigerr = err
        return nil
      }
      err = blob.Unmarshal(b, newfm)
      if err != nil {
        bigerr = err
        return nil
      }
      orig := newfm.Path
      chunks, err = newfm.LoadFromPath(path[1:])
      newfm.Path = orig
      if err != nil {
        bigerr = err
        return nil
      }
    } else {
      var err error
      fmt.Println("new file")
      obj := blob.NewObject()
      newfm.RcasObjectRef = obj.Ref()
      chunks, err = newfm.LoadFromPath(path[1:])
      if err != nil {
        bigerr = err
        return nil
      }
      chunks = append(chunks, obj)
    }

    b, err := blob.Marshal(newfm)
    if err != nil {
      bigerr = err
      return nil
    }
    chunks = append(chunks, b)
    fmt.Println("putting blobs...", len(chunks))

    for _, b := range chunks {
      err := m.Cl.PutBlob(b)
      if err != nil {
        bigerr = err
      }
    }
    return nil
  }

  filepath.Walk("./", fn)
  return bigerr
}

func (m *Mount) runQuery() {
  m.q.Open()
  defer m.q.Close()

  batchN := 1000
  timeout := time.After(10 * time.Second)
  for skip, done := 0, false; !done; skip += batchN {
    blobs, err := m.Cl.BlobsBackward(time.Now(), batchN, skip)
    if len(blobs) > 0 {
      m.q.Process(blobs...)
    }

    if err != nil {
      break
    }
    select {
      case <-timeout:
        done = true
      default:
    }
  }
}

func (m *Mount) Save(pth string) error {
  data, err := json.Marshal(m)
  if err != nil {
    return err
  }

  os.MkdirAll(filepath.Dir(pth), 0744)
  f, err := os.Create(pth)
  if err != nil {
    return err
  }
  defer f.Close()

  f.Write(data)
  return nil
}

func (m *Mount) Load(pth string) error {
  data, err := ioutil.ReadFile(pth)
  if err != nil {
    return err
  }

  err = json.Unmarshal(data, m)
  if err != nil {
    return err
  }
  return nil
}

func (m *Mount) isTip(b *blob.Blob) bool {
  objref := b.ObjectRef()
  tip, err := m.Cl.ObjectTip(objref)
  if err != nil {
    return false
  }
  return b.Ref() == tip.Ref()
}

