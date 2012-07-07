
package main

import (
  "os"
  "io"
  "io/ioutil"
  "fmt"
  "net/http"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
  "encoding/json"
)

const (
  dbServer = "localhost"
)

var (
  db, _ = blobdb.New("./dbase")
)

func main() {
  http.HandleFunc("/", staticHandler)

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

func staticHandler(w http.ResponseWriter, r *http.Request) {
  defer deferWrite(w)

  path := r.URL.Path[1:]
  if path == "cas" {
    static("index.html", w)
  } else if path == "cas/file-upload" {
    static("fupload/index.html", w)
  } else if path == "favicon.ico" {
    static(path, w)
  } else {
    static(path[4:], w)
  }
}

func static(path string, w http.ResponseWriter) {
  f, err := os.Open(path)
  check(err)

  data := make([]byte, 512)
  _, err = f.Read(data)
  check(err)
  w.Header().Set("Content-Type", http.DetectContentType(data))
  _, err = f.Seek(0, 0)
  check(err)

  _, err = io.Copy(w, f)
  check(err)
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
  defer deferPrint()

	mr, err := req.MultipartReader()
  check(err)

  resp := []interface{}{}

	for part, err := mr.NextPart(); err == nil; {
		if name := part.FormName(); name == "" {
      continue
    } else if part.FileName() == "" {
      continue
    }
    fmt.Println("handling file '" + part.FileName() + "'")
    resp = append(resp, storeFileBlob(part.FileName(), part))
		part, err = mr.NextPart()
	}

  data, _ := json.Marshal(resp)
  _, _ = w.Write(data)
}

func storeFileBlob(name string, r io.Reader) (uploadResponse map[string]interface{}) {
  defer func() {
    uploadResponse["name"] = name
    if r := recover(); r != nil {
      uploadResponse["error"] = r.(error).Error()
    }
  }()

  uploadResponse = map[string]interface{}{}

  data, err := ioutil.ReadAll(r)
  check(err)

  uploadResponse["size"] = len(data)

  blobs := blob.SplitRaw(data, blob.DefaultChunkSize)
  refs := blob.RefsFor(blobs)

  meta := blob.NewMeta(blob.FileKind)
  meta.AttachRefs(refs...)
  meta["name"] = name

  b, err := meta.ToBlob()
  check(err)
  err = db.Put(b)
  check(err)
  err = db.Put(blobs...)
  check(err)

  return
}
