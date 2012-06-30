
package main

import (
  "os"
  "io"
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
  http.HandleFunc("/cas", indexHandler)
  http.HandleFunc("/cas/cas.js", jsHandler)
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
  f, err := os.Open("index.html")
  if err != nil {
    fmt.Println(err)
    return
  }
  _, err = io.Copy(w, f)
  if err != nil {
    fmt.Println(err)
  }
}

func jsHandler(w http.ResponseWriter, req *http.Request) {
  w.Header().Set("Content-Type", "text/javascript")

  f, err := os.Open("cas.js")
  if err != nil {
    fmt.Println(err)
    return
  }
  _, err = io.Copy(w, f)
  if err != nil {
    fmt.Println(err)
  }
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
