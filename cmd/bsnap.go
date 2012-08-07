
package main

import (
  "os"
  "fmt"
  "io/ioutil"
  "flag"
  "path/filepath"
  "github.com/rwcarlsen/cas/blobserv"
  "github.com/rwcarlsen/cas/blobdb/query"
  //"code.google.com/p/gopass"
)

var root = flag.String("root", "./", "retrieved file structure is placed here")
var tags = strings.Split(flag.Args(), " ")

var cl = &blobserv.Client{}
func main() {
  home := os.Getenv("HOME")
  confPath := filepath.Join(home, ".blobhostrc")

  conf, err := ioutil.ReadFile(confPath)
  if err != nil {
    fmt.Prinln(err)
    return
  }

  err = json.Unmarshal(conf, cl)
  if err != nil {
    fmt.Prinln(err)
    return
  }

  pth := "./.bmount"
  data, err := ioutil.ReadFile(pth)
  if err != nil {
    fmt.Prinln(err)
    return
  }

  var refs map[string]string
  err := json.Unmarshal(data, &refs)
  if err != nil {
    fmt.Prinln(err)
    return
  }

  fn := func(path string, info os.FileInfo, inerr error) error {
    if info.IsDir() {
      return nil
    }

    var newfm = &blob.FileMeta{}
    var chunks []*blob.Blob
    pth := filepath.Join(path, info.Name())
    if ref, ok := refs[pth]; ok {
      b, err := cl.ObjectTip(ref)
      if err != nil {
        fmt.Println(err)
        return nil
      }
      err = blob.Unmarshal(b, newfm)
      if err != nil {
        fmt.Println(err)
        return nil
      }
      chunks = newfm.LoadFromPath(newfm.Path)
    } else {
      obj := blob.NewObject()
      newfm.RcasObjectRef = obj.Ref()
      chunks = append(newfm.LoadFromPath(newfm.Path), obj)
    }

    b, err := blob.Marshal(newb)
    if err != nil {
      fmt.Println(err)
      return nil
    }
    chunks = append(chunks, b)

    for _, b := range chunks {
      err := cl.PutBlob(b)
      if err != nil {
        fmt.Println(err)
      }
    }

    return nil
  }

  filepath.Walk(".", fn)
  fmt.Println("Snapshot completed.")
}

