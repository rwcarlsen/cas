
package main

import (
  "fmt"
  "flag"
  "github.com/rwcarlsen/cas/mount"
  "github.com/rwcarlsen/cas/util"
)

func main() {
  flag.Parse()
  files := flag.Args()

  if len(files) == 0 {
    files = util.PipedStdin()
  }

  m, err := mount.Load("./.mount")
  if err != nil {
    fmt.Println("Failed to load mount file: ", err)
    return
  }

  for _, path := range files {
    if ref, err := m.GetRef(path); err == nil {
      fmt.Println(ref)
    }
  }
}

