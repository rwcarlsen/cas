
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
  "path/filepath"
)

var (
  DupContentErr = errors.New("blobdb: blob hash-content combo already exist")
  HashCollideErr = errors.New("blobdb: blob hash collision")
)

type dbase struct {
  location string
}

func New(loc string) (db *dbase, err error) {
  var mode os.FileMode = 0744
  if os.MkdirAll(loc, mode); err != nil {
    return nil, err
  }
  return &dbase{location: loc}, nil
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

func (db *dbase) Put(blobs ...*blob.Blob) (err error) {
  // separate loop for error checking makes Puts all or nothing
  var dup error = nil
  for _, b := range blobs {
    ref := b.Ref()
    p := path.Join(db.location, ref)

    if info, err := os.Stat(p); err == nil {
      if info.Size() == int64(len(b.Content)) {
        dup = DupContentErr
      } else {
        return HashCollideErr
      }
    }
  }

  for _, b := range blobs {
    err = db.writeBlob(b)
    if err != nil {
      return err
    }
  }
  return dup
}

// Walk traverses the dbase and returns each blob through the passed 
// channel. Runs in a self-dispatched goroutine
func (db *dbase) Walk() chan *blob.Blob {
  ch := make(chan *blob.Blob)
  fn := func(path string, info os.FileInfo, inerr error) error {
    if info.IsDir() {
      return nil
    }

    b, err := db.Get(path)
    if err != nil {
      return nil
    }
    ch <- b
    return nil
  }

  go filepath.Walk(db.location, fn)

  return ch
}

func (db *dbase) writeBlob(b *blob.Blob) (err error) {
  ref := b.Ref()
  p := path.Join(db.location, ref)
  f, err := os.Create(p)
  if err != nil {
    return err
  }
  defer f.Close()

  _, err = f.Write(b.Content)
  return err
}

func verifyBlob(sum string, b *blob.Blob) (err error) {
  if hex.EncodeToString(b.Sum()) != sum {
    err = errors.New("blobdb: blob name does not match hash of its content.")
  }
  return
}

