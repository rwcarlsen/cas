
package blobdb

import (
  "github.com/rwcarlsen/cas/blob"
)

type FilterFunc func(*blob.Blob) bool

type Filter struct {
  fn FilterFunc

  in chan *blob.Blob
  out chan *blob.Blob

  skip chan *blob.Blob
  done chan bool
}

func (f *Filter) SendTo(other Filter) {
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

func (q *Query) Close() {
  for _, ch := range q.done {
    ch <- true
  }
}

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

