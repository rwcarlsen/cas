
package blob

import (
  "crypto/rand"
)

const (
  randomField = "rcas-random"
)

type object struct {
  RcasType Type
  random []byte
}

// NewObject creates an immutable time-stamped blob that can be used to
// simulate mutable objects that have a dynamic, pruneable revision
// history.
func NewObject() (b *Blob, err error) {
  r := make([]byte, 100)
  if _, err = rand.Read(r); err != nil {
    return nil, err
  }
  o := object {
    RcasType: ObjectType,
    random: r,
  }
  return Marshal(o)
}

