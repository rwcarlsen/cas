
package main

import (
  "fmt"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
)

func main() {
  b := blob.New("sha256")
  b.Write([]byte("hello monkey man"))

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
