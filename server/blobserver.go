
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
)

var (
  db, _ = blobdb.New("./dbase")
)

func main() {
  http.HandleFunc("/get", get)
  http.HandleFunc("/put", putnote)

  fmt.Println("Starting http server...")
  err := http.ListenAndServe("0.0.0.0:7777", nil)

  if err != nil {
    fmt.Println(err)
    return
  }
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

  err := db.Put(b)
  check(err)
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

