
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
var prefix = flag.String("prefix", "", "path prefix that is removed before mounting")

func main() {
  flag.Parse()
  url := flag.Arg(0)

  refs := []string{}
  if len(flag.Args()) > 1 {
    refs = flag.Args()[1:]
  } else {
    piped := util.PipedStdin()
    url = piped[0]
    if len(piped) > 1 {
      refs = piped[1:]
    }
  }

  tmp := strings.Split(url, "@")
  userPass := strings.Split(tmp[0], ":")
  if len(userPass) != 2 || len(tmp) != 2 {
    fmt.Println("Invalid blobserver address")
    return
  }

  m := mount.New(mountPath)
  m.ConfigClient(userPass[0], userPass[1], tmp[1])
  m.Root, _ = filepath.Abs(*root)
  m.Prefix = *prefix

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

func mountPath(f *blob.FileMeta) string {
  mm := &mount.Meta{}
  err := f.GetNotes(mount.Key, mm)

  if *prefix == "" && err != nil {
    return filepath.Join("pathless", f.Name)
  }

  if strings.HasPrefix(mm.Path, *prefix) {
    return filepath.Join(mm.Path[len(*prefix):], f.Name)
  }
  return ""
}

