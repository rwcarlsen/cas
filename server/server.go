
package main

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
  "github.com/rwcarlsen/cas/index"
)

const (
  dbServer = "localhost"
  dbPath = "./dbase"
)

var (
  db, _ = blobdb.New(dbPath)
  ti = index.NewTimeIndex()
)

func main() {
  // initial updating of index
  ch := db.Walk()
  for b := range ch {
    ti.Notify(b)
  }

  http.HandleFunc("/get/", RequireAuth(get))
  http.HandleFunc("/put/", RequireAuth(put))
  http.HandleFunc("/index/", RequireAuth(indexer))
  http.HandleFunc("/share/", RequireAuth(share))

  fmt.Println("Starting http server...")
  err := http.ListenAndServe("0.0.0.0:8888", nil)

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
      m := make(blob.MetaData)
      m["message"] = msg
      resp, _ := blob.Marshal(m)
      w.Write(resp.Content)
    }
  }()
  ref := req.FormValue("ref")

  b, err := db.Get(ref)
  check(err)

  w.Write(b.Content)
}

func put(w http.ResponseWriter, req *http.Request) {
  m := make(blob.MetaData)
  defer func() {
    msg := "blob posted sucessfully"
    if r := recover(); r != nil {
      fmt.Println(r)
      msg = "blob post failed: " + r.(error).Error()
    }

    m["message"] = msg
    resp, _ := blob.Marshal(m)
    w.Write(resp.Content)
  }()

  body, err := ioutil.ReadAll(req.Body)
  check(err)

  b := blob.NewRaw(body)
  m["blob-ref"] = b.Ref()

  err = db.Put(b)
  check(err)
  ti.Notify(b)
}

func indexer(w http.ResponseWriter, req *http.Request) {
  defer deferWrite(w)

  //qname := req.FormValue("query")

  // retrieve query results

  //w.Write(results)
}

func share(w http.ResponseWriter, req *http.Request) {
  defer func() {
    if r := recover(); r != nil {
      fmt.Println(r)
      msg := "blob retrieval failed: " + r.(error).Error()
      m := make(blob.MetaData)
      m["message"] = msg
      resp, _ := blob.Marshal(m)
      w.Write(resp.Content)
    }
  }()

  ref := req.FormValue("ref")
  b, err := db.Get(ref)
  check(err)

  //m := make(blob.MetaData)
  //m, err := blob.Unmarshal(b, &m)
  //check(err)

  //tpe, ok := m[blob.TypeField]
  //if !ok {
  //  // unauthorized
  //  return
  //}

  //if tpe != blob.ShareKind {
  //  // unauthorized
  //  return
  //}

  fname := "what a name"

  head := w.Header()
  head.Set("Content-Type", "application/octet-stream")
  head.Set("Content-Disposition", "attachment; filename=\"" + fname + "\"")
  head.Set("Content-Transfer-Encoding", "binary")
  head.Set("Accept-Ranges", "bytes")
  head.Set("Cache-Control", "private")
  head.Set("Pragma", "private")
  head.Set("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")

  w.Write(b.Content)
}

