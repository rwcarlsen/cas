
package Blob

import (
  "crypto/sha256"
  "hash"
  "encoding/hex"
)

type Blob struct {
  hashFunc hash.Hash
  hashName string
  content []byte
}

func New() *Blob {
  h := sha256.New()
  name := "sha256"
  return &Blob{hashFunc:h, hashName: name}
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

func (b *Blob) FileName() string {
  return b.hashName + "-" + hex.EncodeToString(b.Sum())
}

func (b *Blob) Content() []byte {
  return b.content
}

