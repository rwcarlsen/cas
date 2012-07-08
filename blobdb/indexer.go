
package blobdb

import (
  "github.com/rwcarlsen/cas/blob"
)

type Filter interface {
  Dispatch()
  KillVia(ch chan bool)
  ReceiveVia(ch chan *blob.Blob)
  SendVia(ch chan *blob.Blob)
}

filterFunc

type Query struct {
  filters []Filter
}

func (q *Query) addFilter(f Filter) {
  q.filters = append(q.filters, f)
}

type Result struct {
  in chan *blob.Blob
  done chan bool
  Blobs []*blob.Blob
}

func NewResult() {
  return Result{}
}

func (r *Result) ReceiveVia(ch chan *blob.Blob) {
  r.in = ch
}

func (r *Result) SendVia(ch chan *blob.Blob) { }

func (r *Result) KillVia(ch chan bool) {
  r.done = ch
}

func (r *Result) Dispatch() {
  go func() {
    for {
      select {
        case b := <-r.in:
          r.Blobs = append(r.Blobs, b)
        case <-r.done:
          break
      }
    }
  }()
}

func Pipe(from, to Filter) {
  ch := make(chan *blob.Blob)
  from.SendVia(ch)
  to.ReceiveVia(ch)
}

// filter implementations
type metaFilter struct {
  Field string
  Contains string

  in chan *blob.Blob
  out chan *blob.Blob
}

func NewMetaFilter(q *Query) {

}

func (f *metaFilter) ReceiveVia(ch chan *blob.Blob) {
  f.in = ch
}

func (f *metaFilter) SendVia(ch chan *blob.Blob) {
  f.out = ch
}

func (

