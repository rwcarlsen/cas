
package blobdb 
import (
  "github.com/rwcarlsen/cas/blob"
  "encoding/json"
  "strings"
)

// FilterFunc is used by Filter's to check whether a particular blob can
// pass through or should be skipped/blocked. Returns true for pass
// through.
type FilterFunc func(*blob.Blob) bool

type Filter struct {
  fn FilterFunc

  in chan *blob.Blob
  out chan *blob.Blob

  skip chan *blob.Blob
  done chan bool
}

// SendTo specifies the next filter for blobs that pass through this
// filter. All filters are default initialized to send to a query's
// results.
func (f *Filter) SendTo(other *Filter) {
  f.out = other.in
}

func (f *Filter) dispatch() {
  go func() {
    for {
      select {
        case b := <-f.in:
          if f.fn(b) {
            f.out <- b
          } else {
            f.skip <- b
          }
        case <-f.done:
          return
      }
    }
  }()
}

// Query is used to coordinate arbitrary multi-filter searches through
// blobs.
//
// Note that you can feed a query's own results back into its Process
// method to effectively have dynamic update of time-dependent queries.
type Query struct {
  filters []*Filter
  done []chan bool
  roots []chan *blob.Blob
  skip chan *blob.Blob
  Skipped []*blob.Blob
  Results []*blob.Blob
  result chan *blob.Blob
}

func NewQuery() *Query {
  return &Query{
      filters: make([]*Filter, 0),
      done: make([]chan bool, 0),
      roots: make([]chan *blob.Blob, 0),
      skip: make(chan *blob.Blob),
      result: make(chan *blob.Blob),
      Skipped: make([]*blob.Blob, 0),
      Results: make([]*blob.Blob, 0),
    }
}

func (q *Query) Open() {
  for _, f := range q.filters {
    f.dispatch()
  }
}

// Clear resets a query's Results and Skipped fields (as if no blobs had
// been been processed)
func (q *Query) Clear() {
  q.Skipped = make([]*blob.Blob, 0)
  q.Results = make([]*blob.Blob, 0)
}

// Close terminates and resets the query to blank (i.e. as returned by
// NewQuery).  Neglecting to call Close results in hanging goroutines.
func (q *Query) Close() {
  for _, ch := range q.done {
    ch <- true
  }
  q = NewQuery()
}

// Process passes blobs through the query's filter network and returns when
// all blobs have finished processing.
func (q *Query) Process(blobs ...*blob.Blob) {
  for _, b := range blobs {
    for _, ch := range q.roots {
      ch <- b
      select {
        case res := <-q.result:
          q.Results = append(q.Results, res)
        case sk := <-q.skip:
          q.Skipped = append(q.Skipped, sk)
      }
    }
  }
}

// NewFilter creates a new filter attached to this query.
func (q *Query) NewFilter(fn FilterFunc) *Filter {
  done := make(chan bool)
  q.done = append(q.done, done)
  f := &Filter{
      in: make(chan *blob.Blob),
      fn: fn,
      out: q.result,
      done: done,
      skip: q.skip,
    }
  q.filters = append(q.filters, f)
  return f
}

// SetRoots specifies which filter(s) are the initial receivers of
// processed blobs.
func (q *Query) SetRoots(roots ...*Filter) {
  for _, f := range roots {
    q.roots = append(q.roots, f.in)
  }
}

/////////////////////////////////////////////////////////
/////////////// helpful filter funcs ////////////////////
/////////////////////////////////////////////////////////

func IsJson(b *blob.Blob) bool {
  if err := json.Unmarshal(b.Content, &blob.MetaData{}); err != nil {
    return false
  }
  return true
}

func Contains(substr string) FilterFunc {
  return func(b *blob.Blob) bool {
    s := string(b.Content)
    if strings.Contains(s, substr) {
      return true
    }
    return false
  }
}

func StampedWithin(dur time.Duration) FilterFunc {
  return func(b *blob.Blob) bool {
    var m blob.MetaData
    err := json.Unmarshal(b.Content, &m)
    if err != nil {
      return false
    }

    if val, ok := m[blob.TimeField]; ok {
      t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", val)
      if err != nil {
        return false
      }

      if time.Since(t) > dur {
        return false
      }
      return true
    }
    return false
  }
}

