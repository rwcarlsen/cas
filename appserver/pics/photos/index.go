
package photos

import (
  "time"
  "sync"
)

const IndexType = "photo-index"

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
func (ind *Index) AddPhoto(objref string, p *Photo) {
  ind.lock.Lock()
  defer ind.lock.Unlock()

  for i := len(ind.PhotoRefs) - 1; i > 0; i-- {
    if objref == ind.PhotoRefs[i] {
      return
    }
  }
  ind.PhotoRefs = append(ind.PhotoRefs, objref)
}

