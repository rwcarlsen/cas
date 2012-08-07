
package main

import (
  "os"
  "fmt"
  "io/ioutil"
  "flag"
  "strings"
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
  flag.Parse()
  if strings.HasPrefix(*root, "./") {
    abs, _ := filepath.Abs("./")
    *root = filepath.Join(abs, *root)
  }

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
    if err != nil || !isTip(b) {
      continue
    }

    fm, data, err := cl.ReconstituteFile(b.Ref())
    if err != nil {
      fmt.Println(err)
      continue
    }

    os.MkdirAll(filepath.Join(*root, fm.Path), 0744)
    fmt.Println("root=", *root, ", path=", fm.Path, ", name=", fm.Name)
    fmt.Println("creating file: ", filepath.Join(*root, fm.Path, fm.Name))
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

func isTip(b *blob.Blob) bool {
  objref := b.ObjectRef()
  tip, err := cl.ObjectTip(objref)
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
      fmt.Println("len=", len(blobs))
      fmt.Println("blobs:", blobs)
      q.Process(blobs...)
    }

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
