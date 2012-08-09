
package main

import (
  "fmt"
  "flag"
  "log"
  "github.com/rwcarlsen/cas/mount"
)

var m *mount.Mount
func main() {
  flag.Parse()
  m = mount.New(nil, nil)
  err := m.Load("./.mount")
  if err != nil {
    log.Fatal("Could not find mount configuration")
  }

  files := flag.Args()

  for _, path := range files {
    err := m.Hide(path)
    if err != nil {
      log.Fatal("Coult not remove '" + path + "'")
    }
    fmt.Println("removed '" + path + "'")
  }
}

