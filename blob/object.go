package blob

import (
  "crypto/rand"
)

const (
  randomField = "RcasRandom"
)

type objectMeta struct {
  RcasType string
  RcasRandom []byte
}

// NewObject creates an immutable time-stamped blob that can be used to
// simulate mutable objects that have a dynamic, pruneable revision
// history.
func NewObject() *Blob {
  r := make([]byte, 50)
  if _, err := rand.Read(r); err != nil {
    panic(err)
  }

  o := &objectMeta{}
  o.RcasType = Object
  o.RcasRandom = r

  b, err := Marshal(o)
  if err != nil {
    panic(err)
  }

  return b
}
