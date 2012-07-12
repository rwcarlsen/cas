
package blobdb

import (
  "github.com/rwcarlsen/cas/blob"
  "errors"
  "time"
  "sync"
)

type Indexer struct {
  ti *TimeIndex
  oi *ObjectIndex
}

func NewIndexer() *Indexer {
  return &Indexer{
  }
}

func (ind *Indexer) Notify(blobs ...*blob.Blob) {
  ind.ti.Notify(blobs...)
  //ind.oi.Notify(blobs...)
}

/////////////////////////////////////////////////////////////////////
/////////// Time Index Stuff ////////////////////////////////////////
/////////////////////////////////////////////////////////////////////
type ObjectIndex struct {

}

/////////////////////////////////////////////////////////////////////
/////////// Time Index Stuff ////////////////////////////////////////
/////////////////////////////////////////////////////////////////////
var (
  IndexEndErr = errors.New("blobdb: end of index")
)

type Iter interface {
  Next() (*blob.Blob, error)
  Prev() (*blob.Blob, error)
  Reset()
}

type splitIter struct {
  
}

type TimeEntry struct {
  tm time.Time
  ref string
}

type TimeIndex struct {
  entries []*TimeEntry
  lock sync.RWMutex
}

func NewTimeIndex() *TimeIndex {
  return &TimeIndex{
    entries: make([]*TimeEntry, 0),
  }
}

func (ti *TimeIndex) Notify(blobs ...*blob.Blob) {
  ti.lock.Lock()
  defer ti.lock.Unlock()

  var t time.Time
  for _, b := range blobs {
    m := make(blob.MetaData)
    err := blob.Unmarshal(b, &m)
    if err != nil {
      t = time.Now()
    } else {
      t, err = time.Parse(blob.TimeFormat, m[blob.TimeField].(string))
      if err != nil {
        t = time.Now()
      }
    }

    ti.entries = append(ti.entries, &TimeEntry{tm: t, ref: b.Ref()})
  }
}

func (ti *TimeIndex) Len() int {
  ti.lock.RLock()
  defer ti.lock.RUnlock()
  return len(ti.entries)
}

func (ti *TimeIndex) GetRef(i int) string {
  ti.lock.RLock()
  defer ti.lock.RUnlock()
  return ti.entries[i].ref
}

func (ti *TimeIndex) IndexOf(t time.Time) int {
  ti.lock.RLock()
  defer ti.lock.RUnlock()

  down, up := 0, len(ti.entries)
  curr, done := split(down, up)
  for !done {
    currt := ti.entries[curr].tm

    if t.After(currt) {
      down, up = curr, up
    } else {
      down, up = down, curr
    }

    curr, done = split(down, up)
  }
  return curr
}

func split(prev, curr int) (next int, found bool) {
  if curr - prev <= 1 {
    return 0, true
  }
  return (prev + curr) / 2, false
}

