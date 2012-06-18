
package main

import (
  "fmt"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
)



func main() {
  //testRaw()
  testPointer()
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

func testPointer() {
  db := blobdb.New(".")

  b := blob.Raw([]byte("hello monkey man"))
  p, err := blob.Pointer(b.Ref(), blob.MetaData{"creator":"me", "favorite-cheese": "swiss", "count":4})
  if err != nil {
    fmt.Println(err)
    return
  }

  err = db.Put(b)
  if err != nil {
    fmt.Println(err)
    return
  }

  err = db.Put(p)
  if err != nil {
    fmt.Println(err)
    return
  }

  p2, err := db.Get(p.Ref())
  if err != nil {
    fmt.Println(err)
    return
  }

  fmt.Println(p2)
  fmt.Println(b)
}
