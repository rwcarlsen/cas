
package main

import (
  "os"
  "fmt"
  "io/ioutil"
  "flag"
  "time"
  "encoding/json"
  "path/filepath"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobserv"
  "github.com/rwcarlsen/cas/query"
  //"code.google.com/p/gopass"
)

var root = flag.String("root", "./", "retrieved file structure is placed here")
var tags = flag.Args()

var cl = &blobserv.Client{}
func main() {
  home := os.Getenv("HOME")
  confPath := filepath.Join(home, ".blobhostrc")

  conf, err := ioutil.ReadFile(confPath)
  if err != nil {
    data, _ := json.Marshal(&blobserv.Client{})
    ioutil.WriteFile(confPath, data, 0744)
    fmt.Println("Blank blob host file created")
  }

  err = json.Unmarshal(conf, cl)
  if err != nil {
    fmt.Println(err)
    return
  } else if cl.Host == "" {
    fmt.Println("No blob server host specified in configuration file.")
    return
  }

  q := query.New()
  ft := q.NewFilter(fitsTags)
  q.SetRoots(ft)

  q.Open()
  getAndFilter(q)
  q.Close()

  refs := map[string]string{}
  for _, b := range q.Results {
    fm := &blob.FileMeta{}
    err = blob.Unmarshal(b, fm)
    fmt.Println("creating file from: ", fm)
    if err != nil || !isTip(b, fm) {
      fmt.Println("is not the tip")
      continue
    }

    fm, data, err := cl.ReconstituteFile(b.Ref())

    os.MkdirAll(filepath.Join(*root, fm.Path), 0744)
    f, err := os.Create(filepath.Join(*root, fm.Path, fm.Name))
    if err != nil {
      fmt.Println(err)
      continue
    }
    f.Write(data)
    f.Close()
    refs[filepath.Join(fm.Path, fm.Name)] = fm.RcasObjectRef
  }

  data, err := json.Marshal(refs)
  if err != nil {
    fmt.Println(err)
    return
  }

  os.MkdirAll(*root, 0744)
  pth := filepath.Join(*root, ".bmount")
  f, err := os.Create(pth)
  f.Write(data)
  f.Close()
}

func isTip(b *blob.Blob, fm *blob.FileMeta) bool {
  tip, err := cl.ObjectTip(fm.RcasObjectRef)
  if err != nil {
    return false
  }
  return b.Ref() == tip.Ref()
}

func getAndFilter(q *query.Query) {
  timeout := time.After(10 * time.Second)

  batchN := 1000
  done := false
  for skip := 0; !done; skip += batchN {
    blobs, err := cl.BlobsBackward(time.Now(), batchN, skip)
    if len(blobs) > 0 {
      q.Process(blobs...)
    }

    fmt.Println("nblobsgotten=", len(blobs))
    if err != nil {
      break
    }
    select {
      case <-timeout:
        done = true
      default:
    }
  }
}

func fitsTags(b *blob.Blob) bool {
  f := &blob.FileMeta{}
  err := blob.Unmarshal(b, f)
  if err != nil {
    return false
  }

  items := filepath.SplitList(f.Path)
  for _, tag := range tags {
    matched := false
    for _, item := range items {
      if item == tag {
        matched = true
        break
      }
    }
    if !matched {
      return false
    }
  }
  return true
}
