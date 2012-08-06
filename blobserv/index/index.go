
package index

import (
  "errors"
  "net/http"
  "github.com/rwcarlsen/cas/blob"
)

var (
  IndexEndErr = errors.New("index: end of index")
)

type Index interface {
  Notify(...*blob.Blob)
  GetIter(r *http.Request) (Iter, error)
}

// Iter is used to walk through blob refs of an index.  
//
// When there are no more blobs to iterate over, Next returns an empty string
// along with IndexEndErr
type Iter interface {
  Next() (string, error)
  SkipN(n int)
}

