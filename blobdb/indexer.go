
package blobdb

import (
  "github.com/rwcarlsen/cas/blob"
  "errors"
)

type Indexer struct {
  queries map[string]*Query
}

func NewIndexer() *Indexer {
  return &Indexer{
    queries: make(map[string]*Query, 0),
  }
}

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

// Refresh pipes a query's results back into itself and reprocesses them.
func (ind *Indexer) Refresh(name string) error {
  q, ok := ind.queries[name]
  if !ok {
    return errors.New("blobdb: invalid query name")
  }

  res := q.Results
  q.Clear()
  q.Process(res...)

  return nil
}

func (ind *Indexer) Results(name string) (refs []string, err error) {
  q, ok := ind.queries[name]
  if !ok {
    return nil, errors.New("blobdb: invalid query name")
  }
  return blob.RefsFor(q.Results), nil
}

// NewQuery returns a new query that is automatically bound to this 
// Indexer.
func (ind *Indexer) NewQuery(name string) (q *Query, err error) {
  if _, ok := ind.queries[name]; ok {
    return nil, errors.New("blobdb: query name already exists")
  }
  q = NewQuery()
  ind.queries[name] = q
  return q, nil
}

