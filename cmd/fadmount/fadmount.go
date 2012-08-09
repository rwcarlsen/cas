
package main

import (
  "fmt"
  "flag"
  "strings"
  "path/filepath"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/mount"
  "github.com/rwcarlsen/cas/util"
)

var root = flag.String("root", "./", "retrieved file structure is placed here")
var blobPath = flag.String("path", "", "blobs under path are mounted into root directory")

func main() {
  flag.Parse()
  url := flag.Arg(0)

  refs := []string{}
  if len(flag.Args()) > 1 {
    refs = flag.Args()[1:]
  }

  piped := util.PipedStdin()
  if len(piped) > 1 {
    url = piped[0]
    refs = append(refs, piped[1:]...)
  }

  tmp := strings.Split(url, "@")
  userPass := strings.Split(tmp[0], ":")

  m := mount.New(mountPath)
  m.ConfigClient(userPass[0], userPass[1], tmp[1])
  m.Root, _ = filepath.Abs(*root)
  m.BlobPath = *blobPath

  err := m.Unpack(refs...)
  if err != nil {
    fmt.Println(err)
    return
  }
  err = m.Save(filepath.Join(*root, ".mount"))
  if err != nil {
    fmt.Println(err)
    return
  }
}

func mountPath(fm *blob.FileMeta) string {
  return filepath.Join(fm.Path, fm.Name)[len(*blobPath):]
}

