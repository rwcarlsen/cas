
package main

import (
  "fmt"
  "./blob"
)

func main() {
  b := blob.New()
  b.Write([]byte("hello"))

  fmt.Println(b.FileName())
  fmt.Println(len(b.FileName()))
}
