
package fupload

import (
  "fmt"
  "strings"
  "io/ioutil"
  "net/http"
  "encoding/json"
  "mime/multipart"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/app"
)

func Handle(cx *app.Context, w http.ResponseWriter, r *http.Request) {
  defer util.DeferWrite(w)

  pth := strings.Trim(r.URL.Path, "/")
  if pth == "fupload" {
    err := util.LoadStatic("fupload/index.html", w)
    util.Check(err)
  } else if pth == "fupload/putfiles" {
    putfiles(cx, w, r)
  } else {
    err := util.LoadStatic(pth, w)
    util.Check(err)
  }
}

func putfiles(cx *app.Context, w http.ResponseWriter, req *http.Request) {
  defer util.DeferPrint()

	mr, err := req.MultipartReader()
  util.Check(err)

  resps := []interface{}{}

	for part, err := mr.NextPart(); err == nil; {
		if name := part.FormName(); name == "" {
      continue
    } else if part.FileName() == "" {
      continue
    }
    fmt.Println("handling file '" + part.FileName() + "'")
    resp := sendFileBlobs(cx, part)
    resps = append(resps, resp)
		part, err = mr.NextPart()
	}

  data, _ := json.Marshal(resps)
  w.Write(data)
}

func sendFileBlobs(cx *app.Context, part *multipart.Part) (respMeta blob.MetaData) {
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
  util.Check(err)

  meta.Size = float64(len(data))

  blobs := blob.SplitRaw(data, blob.DefaultChunkSize)
  meta.ContentRefs = blob.RefsFor(blobs)

  m, err := blob.Marshal(meta)
  util.Check(err)

  blobs = append(blobs, m)

  for _, b := range blobs {
    err = cx.PutBlob(b)
    util.Check(err)
  }

  return respMeta
}

