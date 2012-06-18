
package blobdb

import (
  "github.com/rwcarlsen/cas/blob"
  "path"
  "io/ioutil"
  "errors"
  "strings"
  "os"
  "encoding/hex"
  "crypto"
)

type dbase struct {
  location string
}

func New(loc string) *dbase {
  return &dbase{location: loc}
}

func blobRefParts(ref string) (hash crypto.Hash, sum string) {
  parts := strings.Split(ref, blob.NameHashSep)
  if len(parts) != 2 {
    panic("blobdb: Invalref blob ref " + ref)
  }

  return blob.NameToHash(parts[0]), parts[1]
}

func (db *dbase) Get(ref string) (b *blob.Blob, err error) {
  defer func() {
    if r := recover(); r != nil {
      err = r.(error)
    }
  }()

  p := path.Join(db.location, ref)
  f, err := os.Open(p)
  if err != nil {
    return
  }
  defer f.Close()

  data, err := ioutil.ReadAll(f)
  if err != nil {
    return
  }

  hash, sum := blobRefParts(ref)
  b = blob.Raw(data)
  b.Hash = hash

  err = verifyBlob(sum, b)
  return
}

func (db *dbase) Put(b *blob.Blob) (err error) {
  ref := b.Ref()
  p := path.Join(db.location, ref)

  _, err = os.Stat(p)
  if os.IsExist(err) {
    err = errors.New("blobdb: blob " + p + " already exists")
    return
  }

  f, err := os.Create(p)
  if err != nil {
    return
  }
  defer f.Close()

  _, err = f.Write(b.Content)

  return
}

func verifyBlob(sum string, b *blob.Blob) (err error) {
  if hex.EncodeToString(b.Sum()) != sum {
    err = errors.New("blobdb: blob name does not match hash of its content.")
  }
  return
}

