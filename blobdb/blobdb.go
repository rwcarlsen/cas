
package blobdb

import (
  "github.com/rwcarlsen/cas/blob"
  "path"
  "io/ioutil"
  "errors"
  "strings"
)

type dbase struct {
  location string
}

const (
  nameHashSep = "-"
)

func New(loc string) *dbase {
  return &dbase{location: loc}
}

func blobNameParts(id string) (hashName, sum string, err error) {
  err = nil

  parts := strings.Split(id, nameHashSep)
  if len(parts) != 2 {
    err = errors.New("blobdb: Invalid blob id " + id)
    return
  }

  hashName = parts[0]
  sum = parts[1]
  return
}

func (db *dbase) Get(id string) (b *blob.Blob, err error) {
  err = nil

  hashName, sum, err := blobNameParts(id)
  if err != nil {
    return
  }

  b = blob.New(hashName)

  err = verifyBlob(sum, b)
  if err != nil {
    return
  }

  p := path.Join(db.location, id)
  f, err := os.Open(p)
  if err != nil {
    return
  }
  defer f.Close()

  data, err := ioutil.ReadAll(f)
  if err != nil {
    return
  }

  b.Write(data)

  return
}

func (db *dbase) Put(b *blob.Blob) (err error) {
  err = nil
  id := fileName(b)
  p := path.Join(db.location, id)

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

  _, err = f.Write(b.Content())
  if err != nil {
    return
  }

  return
}

func verifyBlob(sum string, b *blob.Blob) (err error) {
  actual := b.Sum()
  err = nil
  if b.Sum() != sum {
    err = errors.New("blobdb: blob name does not match hash of its content.")
  }
  return
}

func FileName(b *blob.Blob) string {
  return b.HashName() + nameHashSep + hex.EncodeToString(b.Sum())
}

