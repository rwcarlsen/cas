
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
  done chan bool
  roots []chan *blob.Blob
  skip chan *blob.Blob
}

func NewQuery() *Query {
  done := make(chan bool)
  skip := make(chan *blob.Blob)
  roots := make([]chan *blob.Blob, 0)
  filters := make([]*Filter, 0)
  return &Query{filters: filters, done: done, skip: skip, roots: roots}
}

func (q *Query) Open() {
  for _, f := range q.filters {
    f.Dispatch()
  }
}

func (q *Query) Close() {
  q.done <- true
}

func (q *Query) Process(blobs ...*blob.Blob) {
  for _, b := range blobs {
    for _, root := range q.roots {
      root <- b
    }
  }
}

func (q *Query) NewFilter(fn FilterFunc) *Filter {
  ch := make(chan *blob.Blob)
  f := &Filter{in: ch, fn: fn, done: q.done, skip: q.skip}
  q.filters = append(q.filters, f)
  return f
}

func (q *Query) NewRootFilter(fn FilterFunc) *Filter {
  ch := make(chan *blob.Blob)
  f := &Filter{in: ch, fn: fn, done: q.done, skip: q.skip}
  q.filters = append(q.filters, f)
  return f
}

