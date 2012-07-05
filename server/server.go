
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
  http.HandleFunc("/cas/get", get)
  http.HandleFunc("/cas/put", put)
  http.HandleFunc("/cas/putfiles/", putfiles)

  fmt.Println("Starting http server...")
  err := http.ListenAndServe("0.0.0.0:8888", nil)

  if err != nil {
    fmt.Println(err)
    return
  }
}

func deferPrint() {
  if r := recover(); r != nil {
    fmt.Println(r)
  }
}

func deferWrite(w http.ResponseWriter) {
  if r := recover(); r != nil {
    fmt.Println(r)
    w.Write([]byte(r.(error).Error()))
  }
}

func check(err error) {
  if err != nil {
    panic(err)
  }
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
  defer deferWrite(w)

  f, err := os.Open("index.html")
  check(err)
  _, err = io.Copy(w, f)
  check(err)
}

func jsHandler(w http.ResponseWriter, req *http.Request) {
  defer deferWrite(w)

  w.Header().Set("Content-Type", "text/javascript")
  f, err := os.Open("cas.js")
  check(err)
  _, err = io.Copy(w, f)
  check(err)
}

func put(w http.ResponseWriter, req *http.Request) {
  defer deferWrite(w)

  body, err := ioutil.ReadAll(req.Body)
  check(err)

  b := blob.Raw(body)
  err = db.Put(b)
  check(err)

  w.Write([]byte(b.String()))
}

func get(w http.ResponseWriter, req *http.Request) {
  defer deferWrite(w)

  ref, err := ioutil.ReadAll(req.Body)
  check(err)

  b, err := db.Get(string(ref))
  check(err)

  w.Write(b.Content)
}

func putfiles(w http.ResponseWriter, req *http.Request) {
  defer deferWrite(w)

	mr, err := req.MultipartReader()
  check(err)

	for part, err := mr.NextPart(); err == nil; {
		if name := part.FormName(); name == "" {
      continue
    } else if part.FileName() == "" {
      continue
    }
    fmt.Println("handling file '" + part.FileName() + "'")
    storeFileBlob(part)
		part, err = mr.NextPart()
	}
	return
}

func storeFileBlob(r io.Reader) {
  data, err := ioutil.ReadAll(r)
  check(err)

  blobs := blob.SplitRaw(data, blob.DefaultChunkSize)
  refs := blob.RefsFor(blobs)

  meta := blob.NewMeta(blob.FileKind)
  meta.AttachRefs(refs...)

  b, err := meta.ToBlob()
  check(err)
  err = db.Put(b)
  check(err)
  err = db.Put(blobs...)
  check(err)
}
