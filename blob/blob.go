
package Blob

import (
  "crypto/sha256"
  "hash"
)

type Blob struct {
  hashFunc hash.Hash
  hashName string
  content []byte
}

func New(hashName string) *Blob {
  var h hash.Hash

  switch hashName {
    case "sha256":
      h = sha256.New()
    default:
      return nil
  }
  return &Blob{hashFunc: h, hashName: hashName}
}

func (b *Blob) Write(data []byte) (n int, err error) {
  n = len(data)
  err = nil

  b.content = append(b.content, data...)
  b.hashFunc.Write(data)

  return
}

func (b *Blob) Sum() []byte {
  sum := b.hashFunc.Sum([]byte{})
  return sum
}

func (b *Blob) Content() []byte {
  return b.content
}

func (b *Blob) HashName() string {
  return b.hashName
}

func (b *Blob) String() string {
  return b.hashName + ":\n" +  string(b.content)
}
