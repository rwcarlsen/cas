
// mount is used to manage mounting of file blobs.

// Blobs can be easily mounted into folders, have updated state snapshot
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
  "github.com/rwcarlsen/cas/util"
)

var UntrackedErr = errors.New("mount: Illegal operation on untracked file")

const Key = "mount"

// Meta contains mount-related meta-information that is stored within each
// Meta's Notes field under Key.
type Meta struct {
  Path string
  Hidden bool
}

type Mount struct {
  Client *blobserv.Client
  Root string // Mounted blobs are placed in this directory.
  Refs map[string]string
  Prefix string
  PathFor func(*blob.Meta)string `json:"-"`
}

// New returns a new mount object with no client configuration.
func New(pathFn func(*blob.Meta)string) *Mount {
  return &Mount{
    PathFor: pathFn,
  }
}

// Load creates a mount object from a saved mount file.
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

// Save persists the mount state/configuration to a file
func (m *Mount) Save(name string) error {
  data, err := json.Marshal(m)
  if err != nil {
    return err
  }

  os.MkdirAll(filepath.Dir(name), 0744)
  f, err := os.Create(name)
  if err != nil {
    return err
  }
  defer f.Close()

  f.Write(data)
  return nil
}

// ConfigClient allows convenient easy setting of the blobserver client info
// associated with this mount.
func (m *Mount) ConfigClient(user, pass, host string) {
  m.Client = &blobserv.Client{
    User: user,
    Pass: pass,
    Host: host,
  }
}

// Unpack mounts files associated with each given ref into the directory
// specified by Root and the associated Meta's mount meta-data.
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

    fm := &blob.Meta{}
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

// Snap creates and sends an object-associated, updated snapshot of a file's
// bytes to the blobserver.
//
// All meta-data associated with the file is left unchanged.
func (m *Mount) Snap(path string) error {
  path = keyClean(path)

  var chunks []*blob.Blob
  newfm, err := m.getTip(path)
  if err == UntrackedErr {
    newfm = blob.NewMeta()
    obj := blob.NewObject()
    newfm.RcasObjectRef = obj.Ref()
    chunks, err = newfm.LoadFromPath(path)
    if err != nil {
      return err
    }

    mpath := filepath.Dir(filepath.Join(m.Prefix, m.keyPath(path)))
    newfm.SetNotes(Key, &Meta{Path: mpath})
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

// GetMeta returns the Mount Notes associated with the given file.
func (m *Mount) GetMeta(pth string) (mm *Meta, err error) {
  defer func() {recover()}()

  ref, err := m.GetRef(pth)
  util.Check(err)
  fm, err := m.getTip(ref)
  util.Check(err)

  mm = &Meta{}
  err = fm.GetNotes(Key, mm)
  if err != nil {
    mm = &Meta{}
  }
  return mm, nil
}

// SetMeta sets the mount Notes associated with the given file.
func (m *Mount) SetMeta(pth string, mm *Meta) (err error) {
  defer func() {recover()}()

  ref, err := m.GetRef(pth)
  util.Check(err)
  fm, err := m.getTip(ref)
  util.Check(err)
  err = fm.SetNotes(Key, mm)
  util.Check(err)

  b, err := blob.Marshal(fm)
  util.Check(err)
  err = m.Client.PutBlob(b)
  util.Check(err)

  return nil
}

// GetRef returns the meta blobref associated with the pth specified file.
func (m *Mount) GetRef(pth string) (string, error) {
  if ref, ok := m.Refs[m.keyPath(pth)]; ok {
    return ref, nil
  }
  return "", errors.New("mount: No tracked file for path '" + pth + "'")
}

// getTip returns the Meta blob for the most recent version of the object
// for which the file specified by path is a part.
func (m *Mount) getTip(path string) (*blob.Meta, error) {
  path = keyClean(path)

  var fm = &blob.Meta{}
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

func (m *Mount) keyPath(pth string) string {
  abs, _ := filepath.Abs(pth)
  rel, _ := filepath.Rel(m.Root, abs)
  return keyClean(rel)
}

func keyClean(path string) string {
  return strings.Trim(path, "./\\")
}

