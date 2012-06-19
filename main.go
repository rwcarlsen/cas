
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

func testFromContent() {
  b := blob.FromContent([]byte("hello monkey man"))
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

  blobs, err := blob.FromFile("foo.txt")
  if err != nil {
    fmt.Println(err)
    return
  }

  err = db.Put(f)
  if err != nil {
    fmt.Println(err)
    return
  }

  err = db.Put(m)
  if err != nil {
    fmt.Println(err)
    return
  }

  m2, err := db.Get(m.Ref())
  if err != nil {
    fmt.Println(err)
    return
  }

  fmt.Println(m2)
  fmt.Println(f)
}
