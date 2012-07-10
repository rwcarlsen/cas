
package blobdb

import (
  "github.com/rwcarlsen/cas/blob"
  "errors"
)

type Indexer struct {
  queries map[string]*Query
}

func (ind *indexer)

func (ind *Indexer) Start() {
  for _, q := range ind.queries {
    q.Open()
  }
}

func (ind *Indexer) Notify(blobs ...*blob.Blob) {
  for _, q := range ind.queries {
    q.Process(blobs...)
  }
}

func (ind *Indexer) Stop() {
  for _, q := range ind.queries {
    q.Close()
  }
}

func (ind *Indexer) RefreshAll() {
  for name, _ := range ind.queries {
    ind.Refresh(name)
  }
}

func (ind *Indexer) Refresh(name string) err {
  q, ok := ind.queries[name]
  if !ok {
    return errors.New("blobdb: invalid query name")
  }

  res := q.Results
  q.Clear()
  q.Process(res...)

  return nil
}

// NewQuery returns a new query that is automatically bound to this 
// Indexer.
func (ind *Indexer) NewQuery(name string) *Query, error {
  if _, ok := ind.queries[name]; ok {
    return nil, errors.New("blobdb: query name already exists")
  }
  q := NewQuery()
  ind.queries[name] = q
  return q, nil
}

