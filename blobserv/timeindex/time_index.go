
package timeindex

import (
  "io/ioutil"
  "encoding/json"
  "errors"
  "time"
  "sync"
  "net/http"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobserv/index"
)

type Direc int

const (
  Backward Direc = iota
  Forward
  Alternating
)

type Request struct {
  Time time.Time
  Dir Direc
  SkipN int
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

func New() *TimeIndex {
  return &TimeIndex{
    entries: make([]*timeEntry, 0),
  }
}

// Notify adds additional blob refs (and their timestamps) to the chronological
// index.
//
//Non-json encoded blobs are ignored by TimeIndex.
func (ti *TimeIndex) Notify(blobs ...*blob.Blob) {
  ti.lock.Lock()
  defer ti.lock.Unlock()

  for _, b := range blobs {
    if b.Type() == blob.NoType {
      continue
    }

    t, err := b.Timestamp()
    if err != nil {
      continue
    }

    ti.entries = append(ti.entries, &timeEntry{tm: t, ref: b.Ref()})
  }
}

func (ti *TimeIndex) Swap(i, j int) {
  ti.entries[i], ti.entries[j] = ti.entries[j], ti.entries[i]
}

func (ti *TimeIndex) Less(i, j int) bool {
  return time.Since(ti.entries[i].tm) > time.Since(ti.entries[j].tm)
}

// GetIter returns an iterator that walks the index according to the
// description in the http request.
func (ti *TimeIndex) GetIter(req *http.Request) (it index.Iter,  err error) {
  data, err := ioutil.ReadAll(req.Body)
  if err != nil {
    return nil, errors.New("timeindex: badly formed query request")
  }

  var r Request
  err = json.Unmarshal(data, &r)
  if err != nil {
    return nil, errors.New("timeindex: badly formed query request")
  }

  switch r.Dir {
    case Forward:
      it = ti.iterForward(r.Time)
    case Backward:
      it = ti.iterBackward(r.Time)
    case Alternating:
      it = ti.iterAround(r.Time)
  }

  it.SkipN(r.SkipN)
  return it, nil
}

// Len returns the number of blob refs in the index.
func (ti *TimeIndex) Len() int {
  ti.lock.RLock()
  defer ti.lock.RUnlock()
  return len(ti.entries)
}

// RefAt returns the blob ref stored at index i (i=0 being the oldest blob)
func (ti *TimeIndex) RefAt(i int) (ref string) {
  defer func() {
    if r := recover(); r != nil {
      ref = ""
    }
  }()
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
  if up < 0 {
    return -1
  }

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

  lowt := ti.entries[down].tm
  upt := ti.entries[up].tm
  lowdiff := int64(time.Since(lowt)) - int64(time.Since(t))
  updiff := int64(time.Since(t)) - int64(time.Since(upt))
  if updiff < lowdiff {
    return up
  }
  return down
}

func split(prev, curr int) (next int, found bool) {
  if curr - prev <= 1 {
    return prev, true
  }
  return (prev + curr) / 2, false
}

// iterAround returns an iterator that starts with the blob created around time t and
// gradually walks outward alternating older-newer.
func (ti *TimeIndex) iterAround(t time.Time) index.Iter {
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
func (ti *TimeIndex) iterForward(t time.Time) index.Iter {
  i := ti.IndexNear(t)
  return &forwardIter{
    at: i,
    ti: ti,
  }
}

// iterBackward returns an iterator that starts with a blob created around time t and
// gradually walks backward in time toward older blobs.
func (ti *TimeIndex) iterBackward(t time.Time) index.Iter {
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
  return "", index.IndexEndErr
}

func (it *forwardIter) SkipN(n int) {
  it.at += n
}

type backwardIter struct {
  at int
  ti *TimeIndex
}

func (it *backwardIter) Next() (ref string, err error) {
  if it.at >= 0 && it.at < it.ti.Len() {
    it.at--
    return it.ti.RefAt(it.at + 1), nil
  }
  return "", index.IndexEndErr
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
    return 0, index.IndexEndErr
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

