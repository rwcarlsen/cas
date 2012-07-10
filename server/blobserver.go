
package main

import (
  "fmt"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
)

const (
  dbServer = "localhost"
  dbPath = "./dbase"
)

var (
  db, _ = blobdb.New(dbPath)
  indexer = blobdb.NewIndexer()
)

func main() {
  ch := db.Walk()

  indexer.Start()
  defer indexer.Stop()

  for b := range ch {
    indexer.Notify(b)
  }

  http.HandleFunc("/get/", RequireAuth(get))
  http.HandleFunc("/put/", RequireAuth(put))
  http.HandleFunc("/index/", RequireAuth(index))

  fmt.Println("Starting http server...")
  err := http.ListenAndServe("0.0.0.0:7777", nil)

  if err != nil {
    fmt.Println(err)
    return
  }
}

func get(w http.ResponseWriter, req *http.Request) {
  defer func() {
    if r := recover(); r != nil {
      fmt.Println(r)
      msg := "blob retrieval failed: " + r.(error).Error()
      m := blob.NewMeta(blob.NoneKind)
      m["message"] = msg
      resp, _ := m.ToBlob()
      w.Write(resp.Content)
    }
  }()

  ref, err := ioutil.ReadAll(req.Body)
  check(err)

  b, err := db.Get(string(ref))
  check(err)

  w.Write(b.Content)
}

func put(w http.ResponseWriter, req *http.Request) {
  m := blob.NewMeta(blob.NoneKind)
  defer func(m blob.MetaData) {
    msg := "blob posted sucessfully"
    if r := recover(); r != nil {
      fmt.Println(r)
      msg = "blob post failed: " + r.(error).Error()
    }

    m["message"] = msg
    resp, _ := m.ToBlob()
    w.Write(resp.Content)
  }(m)

  body, err := ioutil.ReadAll(req.Body)
  check(err)

  b := blob.Raw(body)
  m["blob-ref"] = b.Ref()

  err = db.Put(b)
  check(err)
}

func index(w http.ResponseWriter, req *http.Request) {
  defer deferWrite(w)

  qname, err := ioutil.ReadAll(req.Body)
  check(err)
  refs, err := indexer.Results(string(qname))
  check(err)
  data, err := json.Marshal(refs)
  check(err)

  w.Write(data)
}

