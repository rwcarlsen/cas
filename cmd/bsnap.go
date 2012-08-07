
package main

import (
  "os"
  "fmt"
  "io/ioutil"
  "path/filepath"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/mount"
)

var cl = &blobserv.Client{}
func main() {
  m := mount.New(nil, nil)

  err := m.Load("./.mount")
  if err != nil {
    fmt.Println(err)
    return
  }

  //walk the directory

}

