
package main

import (
  "fmt"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
)



func main() {
  //testRaw()
  //testFile()
  //testDir()
  testIndexer()
}

func testRaw() {
  b := blob.Raw([]byte("hello monkey man"))
  db, _ := blobdb.New(".")

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
  db, _ := blobdb.New(".")

  meta, blobs, err := blob.FileBlobsAndMeta("foodir/foo.txt")
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
  db, _ := blobdb.New(".")

  metas, blobs, err := blob.DirBlobsAndMeta("foodir")
  if err != nil {
    fmt.Println(err)
    return
  }

  metablobs := make([]*blob.Blob, 0)
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

  for _, b := range metablobs {
    fmt.Println(b)
  }
  for _, b := range blobs {
    fmt.Println(b)
  }
}

func testIndexer() {
  b1 := blob.Raw([]byte("I am not json"))
  b2 := blob.Raw([]byte("{\"key\":\"I am json\"}"))

  q := blobdb.NewQuery()
  f := q.NewFilter(blobdb.IsJson)
  q.SetRoots(f)

  q.Open()
  q.Process(b1, b2)
  q.Close()

  fmt.Println("results: ", q.Results)
  fmt.Println("skipped: ", q.Skipped)
}

