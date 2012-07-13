
package index

import (
  "github.com/rwcarlsen/cas/blob"
  "errors"
  "time"
  "sync"
  "net/http"
)

var (
  IndexEndErr = errors.New("index: end of index")
)

type Index interface {
  Notify(...*blob.Blob)
  GetIter(r *http.Request) Iter
}

// Iter is used to walk through blob refs of an index.  
//
// When there are no more blobs to iterate over, Next returns an empty string
// along with IndexEndErr
type Iter interface {
  Next() (string, error)
  SkipN(n int)
}

type timeEntry struct {
  tm time.Time
  ref string
}

// TimeIndex is a thread-safe, iterable, chronological index of blob refs.
type TimeIndex struct {
  entries []*timeEntry
  lock sync.RWMutex
}

func NewTimeIndex() *TimeIndex {
  return &TimeIndex{
    entries: make([]*timeEntry, 0),
  }
}

// Notify adds additional blob refs (and their timestamps) to the chronological
// index.
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

    ti.entries = append(ti.entries, &timeEntry{tm: t, ref: b.Ref()})
  }
}

// Len returns the number of blob refs in the index.
func (ti *TimeIndex) Len() int {
  ti.lock.RLock()
  defer ti.lock.RUnlock()
  return len(ti.entries)
}

// RefAt returns the blob ref stored at index i (i=0 being the oldest blob)
func (ti *TimeIndex) RefAt(i int) string {
  ti.lock.RLock()
  defer ti.lock.RUnlock()
  return ti.entries[i].ref
}

// IndexNear returns the index of the blob created closed to time t. The actual
// blob ref can be retrieved by passing the index to RefAt.
func (ti *TimeIndex) IndexNear(t time.Time) int {
  ti.lock.RLock()
  defer ti.lock.RUnlock()

  down, up := 0, len(ti.entries) - 1
  pivot, done := split(down, up)
  for !done {
    pivt := ti.entries[pivot].tm

    if t.After(pivt) {
      down, up = pivot, up
    } else {
      down, up = down, pivot
    }

    pivot, done = split(down, up)
  }
  return pivot
}

func split(prev, curr int) (next int, found bool) {
  if curr - prev <= 1 {
    return prev, true
  }
  return (prev + curr) / 2, false
}

// GetIter returns an iterator that walks the index according to the
// description in the http request.
func (ti *TimeIndex) GetIter(r *http.Request) Iter {
  // decide how to config iter from request
  // it := iterNew()
  // it := iterOld()
  // it := IterAround()
  // it := IterForward()
  // it := IterBackward()
  return ti.iterNew()
}

// iterNew returns an iterator that starts from the most recent blob ref
// working backward in time.
func (ti *TimeIndex) iterNew() Iter {
  return &backwardIter{
    at: ti.Len() - 1,
    ti: ti,
  }
}

// iterOld returns an iterator that starts from the oldest blob ref working
// forward in time.
func (ti *TimeIndex) iterOld() Iter {
  return &forwardIter{
    at: 0,
    ti: ti,
  }
}

// iterAround returns an iterator that starts with the blob created around time t and
// gradually walks outward alternating older-newer.
func (ti *TimeIndex) iterAround(t time.Time) Iter {
  i := ti.IndexNear(t)
  return &splitIter{
    high: i,
    low: i + 1,
    atTop: true,
    ti: ti,
  }
}

// iterForward returns an iterator that starts with the blob created around time t and
// gradually walks forward in time toward more recent blobs.
func (ti *TimeIndex) iterForward(t time.Time) Iter {
  i := ti.IndexNear(t)
  return &forwardIter{
    at: i,
    ti: ti,
  }
}

// iterBackward returns an iterator that starts with a blob created around time t and
// gradually walks backward in time toward older blobs.
func (ti *TimeIndex) iterBackward(t time.Time) Iter {
  i := ti.IndexNear(t)
  return &backwardIter{
    at: i,
    ti: ti,
  }
}

type forwardIter struct {
  at int
  ti *TimeIndex
}

func (it *forwardIter) Next() (ref string, err error) {
  if it.at < it.ti.Len() {
    it.at++
    return it.ti.RefAt(it.at - 1), nil
  }
  return "", IndexEndErr
}

func (it *forwardIter) SkipN(n int) {
  it.at += n
}

type backwardIter struct {
  at int
  ti *TimeIndex
}

func (it *backwardIter) Next() (ref string, err error) {
  if it.at >= 0 {
    it.at--
    return it.ti.RefAt(it.at + 1), nil
  }
  return "", IndexEndErr
}

func (it *backwardIter) SkipN(n int) {
  it.at -= n
}

type splitIter struct {
  high int
  low int
  atTop bool
  ti *TimeIndex
}

func (it *splitIter) Next() (ref string, err error) {
  i, err := it.next()
  if err != nil {
    return "", err
  }
  return it.ti.RefAt(i), nil
}

func (it *splitIter) next() (i int, err error) {
  low := it.low - 1
  high := it.high + 1

  if low >= 0 && high < it.ti.Len() {
    // both in bounds
    if it.atTop {
      i = low
      it.low--
    } else {
      i = high
      it.high++
    }
    it.atTop = !it.atTop
  } else if low < 0 && high < it.ti.Len() {
    // lower out of bounds
    i = high
    it.high++
  } else if low >= 0 && high >= it.ti.Len() {
    // higher out of bounds
    i = low
    it.low--
  } else {
    // both out of bounds
    return 0, IndexEndErr
  }
  return i, nil
}

func (it *splitIter) SkipN(n int) {
  var err error = nil
  for i := 0; i < n; i++ {
    _, err = it.next()
    if err != nil {
      break
    }
  }
}

