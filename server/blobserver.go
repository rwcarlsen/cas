
package main

import (
  "io/ioutil"
  "fmt"
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

func init() {
  ch := db.Walk()

  indexer.Start()
  defer indexer.Stop()

  for b := range ch {
    indexer.Notify(b)
  }
}

func main() {

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
  m := blob.NewMeta(blob.NoneKind)
  defer func(m blob.MetaData) {
    msg := "blob retrieved sucessfully"
    if r := recover(); r != nil {
      fmt.Println(r)
      msg = "blob retrieval failed: " + r.(error).Error()
    }

    m["message"] = msg
    resp, _ := m.ToBlob()
    w.Write(resp)
  }(m)

  ref, err := ioutil.ReadAll(req.Body)
  check(err)
  m["blob-ref"] = ref

  b, err := db.Get(string(ref))
  check(err)
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
    w.Write(resp)
  }(m)

  body, err := ioutil.ReadAll(req.Body)
  check(err)

  b := blob.Raw(body)
  m["blob-ref"] = b.Ref()

  err = db.Put(b)
  check(err)
}

func index(w http.ResponseWriter, req *http.Request) {
  defer deferWrite()

  qname, err := ioutil.ReadAll(req.Body)
  check(err)
  refs, err := indexer.Results(qname)
  check(err)
  data, err := json.Marshal(refs)
  check(err)

  w.Write(data)
}

