
package photos

import (
  "time"
  "sync"
  "github.com/rwcarlsen/cas/blob"
  "strings"
  "path"
)

const (
  IndexType = "photo-index"
  PhotoType = "photo-meta"
)

type Index struct {
  lock sync.RWMutex
  RcasType string
  RcasObjectRef string
  PhotoRefs []string
  SubIndexes []string
  LastUpdate time.Time
}

func NewIndex() *Index {
  return &Index{
    RcasType: IndexType,
    PhotoRefs: []string{},
    SubIndexes: []string{},
  }
}

func (ind *Index) Newest(n int) (objrefs []string) {
  ind.lock.RLock()
  defer ind.lock.RUnlock()

  l := ind.Len()
  if l <= n {
    return ind.PhotoRefs
  }
  return ind.PhotoRefs[l-n:]
}

func (ind *Index) Len() int {
  return len(ind.PhotoRefs)
}

// AddPhoto adds the objref of a photo to this index
func (ind *Index) AddPhoto(p *Photo) {
  ind.lock.Lock()
  defer ind.lock.Unlock()

  for i := len(ind.PhotoRefs) - 1; i > 0; i-- {
    if p.RcasObjectRef == ind.PhotoRefs[i] {
      return
    }
  }
  ind.PhotoRefs = append(ind.PhotoRefs, p.RcasObjectRef)
}

type Photo struct {
  RcasObjectRef string
  RcasType string
  Who []string
  Tags []string
  Exif map[string]string
  FileObjRef string
  ThumbFileRef string
}

func NewPhoto() *Photo {
  return &Photo{RcasType: PhotoType}
}

func (p *Photo) Copy() *Photo {
  cp := *p

  cp.Who = make([]string, len(p.Who))
  cp.Tags = make([]string, len(p.Tags))
  copy(cp.Who, p.Who)
  copy(cp.Tags, p.Tags)

  for k, v := range p.Exif {
    cp.Exif[k] = v
  }

  return &cp
}

func validImageFile(m *blob.FileMeta) bool {
  if m.RcasType != blob.File {
    return false
  }
  switch strings.ToLower(path.Ext(m.Name)) {
    case ".jpg", ".jpeg", ".gif", ".png": return true
  }
  return false
}
