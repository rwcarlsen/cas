
package main

import (
  "io/ioutil"
  "fmt"
  "net/http"
  //"encoding/json"
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
  http.HandleFunc("/cas", indexHandler)
  http.HandleFunc("/cas/cas.js", casScriptHandler)
  http.HandleFunc("/cas/get", getHandler)
  http.HandleFunc("/cas/put", putHandler)

  fmt.Println("Starting http server...")
  err := http.ListenAndServe("0.0.0.0:8888", nil)
  if err != nil {
    fmt.Println(err)
    return
  }
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
  file_name := "index.html"
  file_data, _ := ioutil.ReadFile(file_name)
  _, _ = w.Write(file_data)
}

func casScriptHandler(w http.ResponseWriter, req *http.Request) {
  file_name := "cas.js"
  file_data, _ := ioutil.ReadFile(file_name)
  w.Header().Set("Content-Type", "text/javascript")
  _, _ = w.Write(file_data)
}

func putHandler(w http.ResponseWriter, req *http.Request) {
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    fmt.Println(err)
    w.Write([]byte(err.Error()))
    return
  }

  b := blob.Raw(body)
  err = db.Put(b)
  if err != nil {
    fmt.Println(err)
    w.Write([]byte(err.Error()))
    return
  }

  w.Write([]byte(b.String()))
}

func getHandler(w http.ResponseWriter, req *http.Request) {
  ref, err := ioutil.ReadAll(req.Body)
  if err != nil {
    fmt.Println(err)
    w.Write([]byte(err.Error()))
    return
  }

  b, err := db.Get(string(ref))
  if err != nil {
    fmt.Println(err)
    w.Write([]byte(err.Error()))
    return
  }
  w.Write(b.Content)
}
