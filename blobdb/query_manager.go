
package blobdb

//func (ind *Indexer) Results(name string) (refs []string, err error) {
//  q, ok := ind.queries[name]
//  if !ok {
//    return nil, errors.New("blobdb: invalid query name")
//  }
//  return q.Results, nil
//}
//
//// NewQuery returns a new query that is automatically bound to this 
//// Indexer.
//func (ind *Indexer) NewQuery(name string) (q *Query, err error) {
//  if _, ok := ind.queries[name]; ok {
//    return nil, errors.New("blobdb: query name already exists")
//  }
//  q = NewQuery()
//  ind.queries[name] = q
//
//  if ind.active {
//    q.Open()
//  }
//  return q, nil
//}

