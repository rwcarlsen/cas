
package recent

import (
  "bytes"
  "time"
  "strings"
  "net/http"
  "encoding/json"
  "html/template"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/app"
  "github.com/rwcarlsen/cas/timeindex"
)

var tmpl *template.Template
func init() {
  tmpl = template.Must(template.ParseFiles("recent/index.tmpl"))
}

func Handle(cx *app.Context, w http.ResponseWriter, r *http.Request) {
  defer util.DeferWrite(w)

  pth := strings.Trim(r.URL.Path, "/")
  if pth == "recent" {
    data := stripBlobs(cx)
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

func stripBlobs(cx *app.Context) []*shortblob {
  indReq := timeindex.Request{
    Time: time.Now(),
    Dir:timeindex.Backward,
  }
  blobs, err := cx.IndexBlobs("time", 20, indReq)
  util.Check(err)

  short := []*shortblob{}
  for _, b := range blobs {
    buf := bytes.NewBuffer([]byte{})
    json.Indent(buf, b.Content(), "", "    ")
    short = append(short, &shortblob{b.Ref(), buf.String()})
  }

  return short
}

