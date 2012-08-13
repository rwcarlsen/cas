
package mount

import (
  "os"
  "io/ioutil"
  "errors"
  "strings"
  "encoding/json"
  "path/filepath"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobserv"
)

var UntrackedErr = errors.New("mount: Illegal operation on untracked file")

type Meta struct {
  Path string
  Hidden bool
}

type Mount struct {
  Client *blobserv.Client
  Root string // Mounted blobs are placed in this directory.
  Refs map[string]string
  Prefix string
  PathFor func(*blob.FileMeta)string `json:"-"`
}

func New(pathFn func(*blob.FileMeta)string) *Mount {
  return &Mount{
    PathFor: pathFn,
  }
}

func Load(pth string) (*Mount, error) {
  data, err := ioutil.ReadFile(pth)
  if err != nil {
    return nil, err
  }

  m := New(nil)
  err = json.Unmarshal(data, m)
  if err != nil {
    return nil, err
  }
  return m, nil
}

func (m *Mount) ConfigClient(user, pass, host string) {
  m.Client = &blobserv.Client{
    User: user,
    Pass: pass,
    Host: host,
  }
}

func (m *Mount) Unpack(refs ...string) error {
  err := m.Client.Dial()
  if err != nil {
    return err
  }

  m.Refs = map[string]string{}
  for _, ref := range refs {
    b, err := m.Client.GetBlob(ref)
    if err != nil {
      return err
    }

    fm := &blob.FileMeta{}
    err = blob.Unmarshal(b, fm)

    fm, data, err := m.Client.ReconstituteFile(b.Ref())
    if err != nil {
      return err
    }

    pth := m.PathFor(fm)
    if pth == "" {
      continue
    }

    full := filepath.Join(m.Root, pth)
    os.MkdirAll(filepath.Dir(full), 0744)
    f, err := os.Create(full)
    if err != nil {
      return err
    }
    f.Write(data)
    f.Close()
    pth = keyClean(pth)
    m.Refs[pth] = ref
  }
  return nil
}

func (m *Mount) GetTip(path string) (*blob.FileMeta, error) {
  path = keyClean(path)

  var fm = &blob.FileMeta{}
  if ref, ok := m.Refs[path]; ok {
    b, err := m.Client.GetBlob(ref)
    if err != nil {
      return nil, err
    }

    b, err = m.Client.ObjectTip(b.ObjectRef())
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
  path = keyClean(path)

  var chunks []*blob.Blob
  newfm, err := m.GetTip(path)
  if err == UntrackedErr {
    newfm = blob.NewFileMeta()
    obj := blob.NewObject()
    newfm.RcasObjectRef = obj.Ref()
    chunks, err = newfm.LoadFromPath(path)
    if err != nil {
      return err
    }

    mpath := filepath.Dir(filepath.Join(m.Prefix, m.keyPath(path)))
    newfm.SetNotes("mount", &Meta{Path: mpath})
    chunks = append(chunks, obj)
  } else if err != nil {
    return err
  }

  chunks, err = newfm.LoadFromPath(path)
  if err != nil {
    return err
  }

  b, err := blob.Marshal(newfm)
  if err != nil {
    return err
  }
  chunks = append(chunks, b)

  m.Refs[path] = b.Ref()

  for _, b := range chunks {
    err := m.Client.PutBlob(b)
    if err != nil {
      return err
    }
  }
  return nil
}

func (m *Mount) keyPath(pth string) string {
  abs, _ := filepath.Abs(pth)
  rel, _ := filepath.Rel(m.Root, abs)
  return keyClean(rel)
}

func (m *Mount) GetRef(pth string) (string, error) {
  if ref, ok := m.Refs[m.keyPath(pth)]; ok {
    return ref, nil
  }
  return "", errors.New("mount: No tracked file for path '" + pth + "'")
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

func keyClean(path string) string {
  return strings.Trim(path, "./\\")
}

