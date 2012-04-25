
package main

import (
  "fmt"
  "cas/blob"
  "cas/blobdb"
)

func main() {
  b := blob.New()
  b.Write([]byte("hello monkey man"))

  db := blobdb.New(".")

  err := db.Put(b)
  if err != nil {
    fmt.Println(err)
  }

  b2, err := db.Get(blobdb.FileName(b))

  fmt.Println(b)
  fmt.Println(b2)
}
