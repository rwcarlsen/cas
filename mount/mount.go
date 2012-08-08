
package mount

import (
  "os"
  "io/ioutil"
  "errors"
  "time"
  "strings"
  "encoding/json"
  "path/filepath"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobserv"
  "github.com/rwcarlsen/cas/query"
)

var UntrackedErr = errors.New("mount: Illegal operation on untracked file")

type Mount struct {
  Client *blobserv.Client
  Root string // Mounted blobs are placed in this directory.
  BlobPath string // All blobs under this meta-path will be mounted.
  Refs map[string]string
  UseHistory bool // false to mount only the most recent version of each blob.
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
  m.Client = &blobserv.Client{
    User: user,
    Pass: pass,
    Host: host,
  }
}

func (m *Mount) Unpack() error {
  err := m.Client.Dial()
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

    fm, data, err := m.Client.ReconstituteFile(b.Ref())
    if err != nil {
      return err
    }

    if fm.Hidden {
      continue
    }

    pth := filepath.Join(m.Root, m.PathFor(fm))
    os.MkdirAll(filepath.Dir(pth), 0744)
    f, err := os.Create(pth)
    if err != nil {
      return err
    }
    f.Write(data)
    f.Close()
    pth = m.PathFor(fm)
    pth = strings.Trim(pth, "./\\")
    m.Refs[pth] = fm.RcasObjectRef
  }
  return nil
}

func (m *Mount) Hide(path string) error {
  fm, err := m.GetTip(path)
  if err != nil {
    return errors.New("mount: Failed to retrieve file meta blob for '" + path + "'")
  }

  fm.Hidden = true

  b, err := blob.Marshal(fm)
  if err != nil {
    return errors.New("mount: Failed to marshal file meta blob")
  }

  err = m.Client.PutBlob(b)
  if err != nil {
    return errors.New("mount: Could not send blob to blobserver")
  }
  return nil
}

func (m *Mount) GetTip(path string) (*blob.FileMeta, error) {
  path = strings.Trim(path, "./\\")

  var fm = &blob.FileMeta{}
  if ref, ok := m.Refs[path]; ok {
    b, err := m.Client.ObjectTip(ref)
    if err != nil {
      return nil, err
    }
    err = blob.Unmarshal(b, fm)
    if err != nil {
      return nil, err
    }
    return fm, nil
  }
  return nil, UntrackedErr
}

func (m *Mount) Snap(path string) error {
  path = strings.Trim(path, "./\\")

  var chunks []*blob.Blob
  newfm, err := m.GetTip(path)
  if err == UntrackedErr {
    newfm = blob.NewFileMeta()
    obj := blob.NewObject()
    m.Refs[path] = obj.Ref()
    newfm.RcasObjectRef = obj.Ref()
    chunks, err = newfm.LoadFromPath(path)
    if err != nil {
      return err
    }
    rel, _ := filepath.Rel(m.Root, newfm.Path)
    newfm.Path = filepath.Join(m.BlobPath, rel)
    chunks = append(chunks, obj)
  } else if err != nil {
    return err
  }

  orig := newfm.Path
  chunks, err = newfm.LoadFromPath(path)
  newfm.Path = orig
  if err != nil {
    return err
  }

  b, err := blob.Marshal(newfm)
  if err != nil {
    return err
  }
  chunks = append(chunks, b)

  for _, b := range chunks {
    err := m.Client.PutBlob(b)
    if err != nil {
      return err
    }
  }
  return nil
}

func (m *Mount) runQuery() {
  m.q.Open()
  defer m.q.Close()

  batchN := 1000
  timeout := time.After(10 * time.Second)
  for skip, done := 0, false; !done; skip += batchN {
    blobs, err := m.Client.BlobsBackward(time.Now(), batchN, skip)
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
  tip, err := m.Client.ObjectTip(objref)
  if err != nil {
    return false
  }
  return b.Ref() == tip.Ref()
}

