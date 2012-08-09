
package main

import (
  "fmt"
  "flag"
  "strings"
  "path/filepath"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/query"
  "github.com/rwcarlsen/cas/mount"
)

var root = flag.String("root", "./", "retrieved file structure is placed here")
var blobPath = flag.String("path", "", "blobs under path are mounted into root directory")

func main() {
  flag.Parse()
  url := flag.Arg(0)
  tmp := strings.Split(url, "@")
  userPass := strings.Split(tmp[0], ":")

  q := query.New()
  ft := q.NewFilter(filtFn(*blobPath))
  q.SetRoots(ft)

  m := mount.New(mountPath, q)
  m.ConfigClient(userPass[0], userPass[1], tmp[1])
  m.Root, _ = filepath.Abs(*root)
  m.BlobPath = *blobPath

  err := m.Unpack()
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

func filtFn(prefix string) func(*blob.Blob)bool {
  return func(b *blob.Blob) bool {
    f := &blob.FileMeta{}
    err := blob.Unmarshal(b, f)
    if err != nil {
      return false
    }
    return strings.HasPrefix(f.Path, strings.Trim(prefix, "./\\"))
  }
}
