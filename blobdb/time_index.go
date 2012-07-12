
package blobdb

import (
  "github.com/rwcarlsen/cas/blob"
  "errors"
  "time"
  "sync"
)

var (
  IndexEndErr = errors.New("blobdb: end of index")
)

type Iter interface {
  Next() (string, error)
  SkipN(n int)
}

type forwardIter struct {
  at int
  ti *TimeIndex
}

func (it *forwardIter) Next() (ref string, err error) {
  if it.at < it.ti.Len() {
    it.at++
    return it.ti.GetRef(it.at - 1), nil
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
    return it.ti.GetRef(it.at + 1), nil
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
  return it.ti.GetRef(i), nil
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


type TimeEntry struct {
  tm time.Time
  ref string
}

// TimeIndex is a thread-safe, iterable, chronological index of blob refs.
type TimeIndex struct {
  entries []*TimeEntry
  lock sync.RWMutex
}

func NewTimeIndex() *TimeIndex {
  return &TimeIndex{
    entries: make([]*TimeEntry, 0),
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

    ti.entries = append(ti.entries, &TimeEntry{tm: t, ref: b.Ref()})
  }
}

// Len returns the number of blob refs in the index.
func (ti *TimeIndex) Len() int {
  ti.lock.RLock()
  defer ti.lock.RUnlock()
  return len(ti.entries)
}

// GetRef returns the blob ref stored at index i (i=0 being the oldest blob)
func (ti *TimeIndex) GetRef(i int) string {
  ti.lock.RLock()
  defer ti.lock.RUnlock()
  return ti.entries[i].ref
}

// IterRecent returns an iterator that starts from the most recent blob ref
// working backward in time.
func (ti *TimeIndex) IterRecent() Iter {
  return &backwardIter{
    at: ti.Len() - 1,
    ti: ti,
  }
}

// IterOld returns an iterator that starts from the oldest blob ref working
// forward in time.
func (ti *TimeIndex) IterOld() Iter {
  return &forwardIter{
    at: 0,
    ti: ti,
  }
}

// IterFrom returns an iterator that starts with the blob created closest to time t and
// gradually walks outward alternating older-newer.
func (ti *TimeIndex) IterFrom(t time.Time) Iter {
  i := ti.IndexNear(t)
  return &splitIter{
    high: i,
    low: i + 1,
    atTop: true,
    ti: ti,
  }
}

// IndexNear returns the index of the blob created closed to time t. The actual
// blob ref can be retrieved by passing the index to GetRef.
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

