
package blob

import (
  "crypto/rand"
)

const (
  randomField = "RcasRandom"
)

type objectMeta struct {
  Type string
  RcasRandom []byte
}

// NewObject creates an immutable time-stamped blob that can be used to
// simulate mutable objects that have a dynamic, pruneable revision
// history.
func NewObject() (b *Blob, err error) {
  r := make([]byte, 100)
  if _, err = rand.Read(r); err != nil {
    return nil, err
  }

  o := &objectMeta{}
  o.Type = Object
  o.RcasRandom = r

  return Marshal(o)
}
