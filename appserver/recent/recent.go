
package recent

import (
  "bytes"
  "time"
  "strings"
  "net/http"
  "encoding/json"
  "html/template"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/blobserv"
)

var tmpl *template.Template
func init() {
  tmpl = template.Must(template.ParseFiles("recent/index.tmpl"))
}

func Handle(c *blobserv.Client, w http.ResponseWriter, r *http.Request) {
  defer util.DeferWrite(w)

  pth := strings.Trim(r.URL.Path, "/")
  if pth == "recent" {
    data := stripBlobs(c)
    err := tmpl.Execute(w, data)
    util.Check(err)
  } else {
    err := util.LoadStatic(pth, w)
    util.Check(err)
  }
}

type shortblob struct {
  Ref string
  Content string
}

func stripBlobs(c *blobserv.Client) []*shortblob {
  blobs, err := c.BlobsBackward(time.Now(), 20, 0)
  util.Check(err)

  short := []*shortblob{}
  for _, b := range blobs {
    buf := bytes.NewBuffer([]byte{})
    json.Indent(buf, b.Content(), "", "    ")
    short = append(short, &shortblob{b.Ref(), buf.String()})
  }

  return short
}

