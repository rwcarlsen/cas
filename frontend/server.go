
package main

import (
  "os"
  "io"
  "io/ioutil"
  "fmt"
  "bytes"
  "net/url"
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

  note["RcasType"] = blob.NoteType

  b, err := blob.Marshal(note)
  check(err)

  addr := hostURL(req).String()
  resp, err := http.Post(addr, "application/octed-stream", bytes.NewBuffer(b.Content))
  resp.Body.Close()
  check(err)

  respData, err := json.MarshalIndent(note, "", "    ")
  check(err)

  w.Write(respData)
}

func putfiles(w http.ResponseWriter, req *http.Request) {
  defer deferPrint()

	mr, err := req.MultipartReader()
  check(err)

  resps := []interface{}{}

  addr := hostURL(req).String()

	for part, err := mr.NextPart(); err == nil; {
		if name := part.FormName(); name == "" {
      continue
    } else if part.FileName() == "" {
      continue
    }
    fmt.Println("handling file '" + part.FileName() + "'")
    resp := sendFileBlobs(part, addr)
    resps = append(resps, resp)
		part, err = mr.NextPart()
	}

  data, _ := json.Marshal(resps)
  w.Write(data)
}

func sendFileBlobs(part *multipart.Part, addr string) (respMeta blob.MetaData) {
  meta := blob.NewFileMeta()
  defer func() {
    data, _ := json.Marshal(meta)
    json.Unmarshal(data, &respMeta)
    delete(respMeta, "ContentRefs")

    if r := recover(); r != nil {
      respMeta["error"] = r.(error).Error()
    }
  }()

  meta.Name = part.FileName()

  data, err := ioutil.ReadAll(part)
  check(err)

  meta.Size = int64(len(data))

  blobs := blob.SplitRaw(data, blob.DefaultChunkSize)
  meta.ContentRefs = blob.RefsFor(blobs)

  m, err := blob.Marshal(meta)
  check(err)

  blobs = append(blobs, m)

  for _, b := range blobs {
    resp, err := http.Post(addr, "application/octed-stream",bytes.NewBuffer(b.Content))
    resp.Body.Close()
    check(err)
  }

  return respMeta
}

func get(w http.ResponseWriter, req *http.Request) {
  defer deferWrite(w)

  addr := hostURL(req)
  resp, err := http.Get(addr.String())
  check(err)

  _, err = io.Copy(w, resp.Body)
  resp.Body.Close()
  check(err)
}

func hostURL(r *http.Request) *url.URL {
  err := r.ParseForm()
  check(err)
  return &url.URL{Host: r.Form.Get("BlobServerHost"), RawQuery: r.Form.Encode()}
}

