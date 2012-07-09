
package blobdb

import (
  "github.com/rwcarlsen/cas/blob"
)

type FilterFunc func(*blob.Blob) bool

type Filter struct {
  fn FilterFunc

  in chan *blob.Blob
  out chan *blob.Blob
  done chan bool
}

func (f *Filter) takeFrom(ch chan *blob.Blob) {
  f.in = ch
}

func (f *Filter) SendTo(other Filter) {
  other.takeFrom(f.out)
}

func (f *Filter) Dispatch() {
  go func() {
    for {
      select {
        case b := <-f.in:
          if f.fn(b) {
            f.out <- b
          }
        case <-f.done:
          return
      }
    }
  }()
}

type Query struct {
  filters []*Filter
  done chan bool
}

func NewQuery() *Query {
  done := make(chan bool)
  filters := make([]*Filter, 0)
  return &Query{filters: filters, done: done}
}

func (q *Query) NewFilter(fn FilterFunc) *Filter {
  ch := make(chan *blob.Blob)
  f := &Filter{in: ch, fn: fn, done: q.done}
  q.filters = append(q.filters, f)
  return f
}

func (q *Query) Run() {
  for _, f := range q.filters {
    f.Dispatch()
  }
  // create intelligent way to terminate all filters
  //q.done <- true
}

