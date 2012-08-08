
package main

import (
  "fmt"
  "github.com/rwcarlsen/cas/mount"
)

func main() {
  m := mount.New(nil, nil)
  err := m.Load("./.mount")
  if err != nil {
    fmt.Println(err)
    return
  }

  err = m.Snapshot()
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Println("Snapshot completed.")
}

