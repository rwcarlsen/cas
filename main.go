
package main

import (
  "fmt"
  "./blob"
  "encoding/hex"
)

func main() {
  b := blob.New()
  b.Write([]byte("hello"))

  fmt.Println(hex.EncodeToString(b.Sum()))
}
