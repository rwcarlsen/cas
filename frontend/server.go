
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
  "mime/multipart"
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
  http.HandleFunc("/cas/putnote", putnote)
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

  pth := r.URL.Path[1:]
  if pth == "cas/" {
    static("index.html", w)
  } else if pth == "cas/file-upload" {
    static("fupload/index.html", w)
  } else if pth == "cas/note-drop" {
    static("notedrop/index.html", w)
  } else if pth == "favicon.ico" {
    static(pth, w)
  } else {
    if len(pth) > 4 {
      static(pth[4:], w)
    }
  }
}

func static(pth string, w http.ResponseWriter) {
  f, err := os.Open(pth)
  check(err)

  w.Header().Set("Content-Type", contentType(pth, f))

  _, err = io.Copy(w, f)
  check(err)
}

func putnote(w http.ResponseWriter, req *http.Request) {
  defer deferWrite(w)

  body, err := ioutil.ReadAll(req.Body)
  check(err)

  var note blob.MetaData
  err = json.Unmarshal(body, &note)
  check(err)

  meta := blob.NewMeta(blob.NoteKind)
  for key, val := range meta {
    note[key] = val
  }

  b, err := note.ToBlob()
  check(err)
  err = db.Put(b)
  check(err)

  resp, err := json.MarshalIndent(note, "", "    ")
  check(err)

  w.Write(resp)
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
    resp = append(resp, storeFileBlob(part))
		part, err = mr.NextPart()
	}

  data, _ := json.Marshal(resp)
  _, _ = w.Write(data)
}

func storeFileBlob(part *multipart.Part) (meta blob.MetaData) {
  defer func() {
    delete(meta, "refs")
    if r := recover(); r != nil {
      meta["error"] = r.(error).Error()
    }
  }()

  meta = blob.NewMeta(blob.FileKind)
  meta["name"] = part.FileName()

  data, err := ioutil.ReadAll(part)
  check(err)

  meta["size"] = len(data)

  blobs := blob.SplitRaw(data, blob.DefaultChunkSize)
  refs := blob.RefsFor(blobs)
  meta.AttachRefs(refs...)

  m, err := meta.ToBlob()
  check(err)

  err = db.Put(m)
  if err != blobdb.DupContentErr {
    check(err)
  }

  err = db.Put(blobs...)
  check(err)

  return
}

func get(w http.ResponseWriter, req *http.Request) {
  defer deferWrite(w)

  ref, err := ioutil.ReadAll(req.Body)
  check(err)

  b, err := db.Get(string(ref))
  check(err)

  w.Write(b.Content)
}

