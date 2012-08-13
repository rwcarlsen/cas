
package main

import (
  "fmt"
  "flag"
  "github.com/rwcarlsen/cas/mount"
  "github.com/rwcarlsen/cas/util"
)

var hide = flag.Bool("hide", false, "fadfind -hidden flag ignores if true")
var pth = flag.String("path", "NOPATH", "path is used to determine file mount location")

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
    if mm, err := m.GetMeta(path); err == nil {
      if *hide {
        mm.Hidden = true
      }
      if *pth != "NOPATH" {
        mm.Path = *pth
      }
      err = m.SetMeta(path, mm)
      if err != nil {
        fmt.Println("Could not update file '" + path + "'")
      }
    }
  }
}

