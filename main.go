
package main

import (
  "fmt"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
)



func main() {
  //testFromContent()
  //testPointer()
  testFile()
}

func testRaw() {
  b := blob.Raw([]byte("hello monkey man"))
  db := blobdb.New(".")


  err := db.Put(b)
  if err != nil {
    fmt.Println(err)
    return
  }

  b2, err := db.Get(b.Ref())
  if err != nil {
    fmt.Println(err)
    return
  }

  fmt.Println(b)
  fmt.Println(b2)
}

func testFile() {
  db := blobdb.New(".")

  meta, blobs, err := blob.FileBlobsAndMeta("foo.txt")
  if err != nil {
    fmt.Println(err)
    return
  }

  err = db.Put(blobs...)
  if err != nil {
    fmt.Println(err)
    return
  }

  m, _ := meta.ToBlob()
  err = db.Put(m)
  if err != nil {
    fmt.Println(err)
    return
  }

  for _, b := range blobs {
    fmt.Println(b)
  }
}

func testDir() {
  db := blobdb.New(".")

  metas, blobs, err := blob.DirBlobsAndMeta("foodir")
  if err != nil {
    fmt.Println(err)
    return
  }

  metablobs := make([]MetaData, 0)
  for _, meta := range metas {
    m, _ := meta.ToBlob()
    metablobs = append(metablobs, m)
  }

  err = db.Put(metablobs...)
  if err != nil {
    fmt.Println(err)
    return
  }

  err = db.Put(blobs...)
  if err != nil {
    fmt.Println(err)
    return
  }

  for _, b := range blobs {
    fmt.Println(b)
  }
}

