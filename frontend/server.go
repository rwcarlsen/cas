
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
  http.HandleFunc("/get", get)
  http.HandleFunc("/putnote", putnote)
  http.HandleFunc("/putfiles/", putfiles)

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
  if pth == "" {
    static("index.html", w)
  } else if pth == "file-upload" {
    static("fupload/index.html", w)
  } else if pth == "note-drop" {
    static("notedrop/index.html", w)
  } else {
    static(pth, w)
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

  // build and send request to blobserver
  host := hostString(req)
  client := &http.Client{}
  r, err := http.NewRequest("POST", host, bytes.NewBuffer(b.Content))
  check(err)
  r.URL.Path = "/put/"
  setAuth(r)
  _, err = client.Do(r)
  check(err)

  buf := bytes.NewBuffer([]byte{})
  err = json.Indent(buf, b.Content, "", "    ")
  check(err)

  w.Write(buf.Bytes())
}

func setAuth(r *http.Request) {
  r.SetBasicAuth("robert", "password")
}

func hostString(r *http.Request) string {
  h := r.Header.Get("Blob-Server-Host")
  // stupid default blob server (temporary)
  if h == "" {
    h = "rwc-server.dyndns.org:7777"
  }
  u := &url.URL{Host: h, Scheme: "http"}
  return u.String()
}

func putfiles(w http.ResponseWriter, req *http.Request) {
  defer deferPrint()

	mr, err := req.MultipartReader()
  check(err)

  resps := []interface{}{}

  host := hostString(req)

	for part, err := mr.NextPart(); err == nil; {
		if name := part.FormName(); name == "" {
      continue
    } else if part.FileName() == "" {
      continue
    }
    fmt.Println("handling file '" + part.FileName() + "'")
    resp := sendFileBlobs(part, host)
    resps = append(resps, resp)
		part, err = mr.NextPart()
	}

  data, _ := json.Marshal(resps)
  w.Write(data)
}

func sendFileBlobs(part *multipart.Part, host string) (respMeta blob.MetaData) {
  meta := blob.NewFileMeta()
  defer func() {
    respMeta = make(blob.MetaData)
    respMeta["name"] = meta.Name
    respMeta["size"] = meta.Size

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

  client := &http.Client{}
  for _, b := range blobs {
    r, err := http.NewRequest("POST", host, bytes.NewBuffer(b.Content))
    check(err)
    r.URL.Path = "/put/"
    setAuth(r)
    _, err = client.Do(r)
    check(err)
  }

  return respMeta
}

func get(w http.ResponseWriter, req *http.Request) {
  defer deferWrite(w)

  // build and send request to blobserver
  host := hostString(req)
  client := &http.Client{}
  r, err := http.NewRequest("GET", host, nil)
  check(err)
  r.URL.Path = "/get/"
  err = req.ParseForm()
  check(err)
  r.Form = req.Form
  setAuth(r)
  resp, err := client.Do(r)
  check(err)

  _, err = io.Copy(w, resp.Body)
  resp.Body.Close()
  check(err)
}

