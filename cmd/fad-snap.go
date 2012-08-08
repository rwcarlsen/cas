
package main

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"
  "github.com/rwcarlsen/cas/mount"
)

func main() {
  m := mount.New(nil, nil)
  err := m.Load("./.mount")
  if err != nil {
    fmt.Println(err)
    return
  }

  fn := func(path string, info os.FileInfo, inerr error) error {
    if info.IsDir() || strings.HasSuffix(path, ".mount") {
      return nil
    }

    err := m.Snap(path)
    if err != nil {
      fmt.Println(err)
    }
    return nil
  }

  filepath.Walk("./", fn)
  fmt.Println("Snapshot completed.")
}

