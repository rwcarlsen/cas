
package objindex

import (
  "encoding/json"
  "io/ioutil"
  "time"
  "sort"
  "errors"
  "sync"
  "net/http"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/index"
)

type Request struct {
  ObjectRef string
  SkipN int
}

type object struct {
  versions []string
  tms []time.Time
}

func (o *object) Add(b *blob.Blob) {
  o.versions = append(o.versions, b.Ref())
  t, err := b.Timestamp()
  util.Check(err)
  o.tms = append(o.tms, t)
}

func (o *object) Swap(i, j int) {
  o.versions[i], o.versions[j] = o.versions[j], o.versions[i]
}

func (o *object) Less(i, j int) bool {
  return time.Since(o.tms[i]) > time.Since(o.tms[j])
}

func (o *object) Len() int {
  return len(o.versions)
}

func (o *object) At(i int) string {
  return o.versions[i]
}

type ObjectIndex struct {
  objs map[string]*object
  lock sync.RWMutex
}

func New() *ObjectIndex {
  return &ObjectIndex{
    objs: map[string]*object{},
  }
}

// Notify adds additional blob refs to the object index if they have an
// object ref.
//
//Blobs larger than MaxBlobSize and non-json encoded blobs are ignored
// by TimeIndex
func (ind *ObjectIndex) Notify(blobs ...*blob.Blob) {
  ind.lock.Lock()
  defer ind.lock.Unlock()

  for _, b := range blobs {
    oref := b.ObjectRef()
    if oref == "" {
      continue
    }

    if ind.objs[oref] == nil {
      ind.objs[oref] = &object{}
    }

    ind.objs[oref].Add(b)
  }
}

// GetIter returns an iterator that walks the index according to the
// description in the http request.
func (ind *ObjectIndex) GetIter(req *http.Request) (it index.Iter,  err error) {
  data, err := ioutil.ReadAll(req.Body)
  if err != nil {
    return nil, errors.New("objindex: badly formed query request")
  }

  var r Request
  err = json.Unmarshal(data, &r)
  if err != nil {
    return nil, errors.New("objindex: badly formed query request")
  }

  it = newIter(ind, r.ObjectRef)
  it.SkipN(r.SkipN)
  return it, nil
}

// Sort organizes the refs for each object in chronological order.
//
// Use this to properly establish an object index that has just been
// initialized by blobs not passed in chronological order.
func (ind *ObjectIndex) Sort() {
  for _, obj := range ind.objs {
    sort.Sort(obj)
  }
}

func (ind *ObjectIndex) RefAt(objref string, i int) (ref string,  err error) {
  ind.lock.RLock()
  defer ind.lock.RUnlock()

  if _, ok := ind.objs[objref]; !ok {
    return "", errors.New("objindex: invalid object ref")
  }
  return ind.objs[objref].At(i), nil
}

// Len returns the number of blob refs in the index.
func (ind *ObjectIndex) ObjLen(objref string) int {
  ind.lock.RLock()
  defer ind.lock.RUnlock()

  if _, ok := ind.objs[objref]; !ok {
    return 0
  }
  return ind.objs[objref].Len()
}

// Len returns the number of blob refs in the index.
func (ind *ObjectIndex) Len() int {
  ind.lock.RLock()
  defer ind.lock.RUnlock()

  tot := 0
  for _, obj := range ind.objs {
    tot += obj.Len()
  }
  return tot
}

type iter struct {
  at int
  objref string
  ind *ObjectIndex
}

func newIter(ind *ObjectIndex, objref string) *iter {
  size := ind.ObjLen(objref)
  return &iter{
    ind: ind,
    at: size - 1,
    objref: objref,
  }
}

func (it *iter) Next() (ref string, err error) {
  if it.at >= 0 && it.at < it.ind.ObjLen(it.objref) {
    it.at--
    return it.ind.RefAt(it.objref, it.at + 1)
  }
  return "", index.IndexEndErr
}

func (it *iter) SkipN(n int) {
  it.at -= n
}

