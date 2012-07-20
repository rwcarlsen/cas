
package recent

import (
  "strings"
  "io/ioutil"
  "net/http"
  "encoding/json"
  "html/template"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/app"
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
    tmpl.Execute(w, data)
  } else {
    err := util.LoadStatic(pth, w)
    util.Check(err)
  }
}

type blob struct {
  Ref string
  Content string
}

func stripBlobs(cx *app.Context) []*blob {
  indReq := timeindex.Request{
    Time: time.Now(),
    Dir:timeindex.Backward,
  }
  blobs, err := cx.IndexBlobs("time", 20, indReq)
  util.Check(err)

  short := []*blob{}
  for _, b := blobs {
    buf := bytes.NewBuffer([]byte{})
    json.Indent(buf, b.Content, "", "    ")
    content := template.HTMLEscapeString(buf.String())
    short = append(short, &blob{b.Ref(), content})
  }

  return short
}

func putnote(cx *app.Context, w http.ResponseWriter, req *http.Request) {
  defer util.DeferWrite(w)

  body, err := ioutil.ReadAll(req.Body)
  util.Check(err)

  var note blob.MetaData
  err = json.Unmarshal(body, &note)
  util.Check(err)

  note["RcasType"] = blob.NoteType

  b, err := blob.Marshal(note)
  util.Check(err)

  err = cx.PutBlob(b)
  util.Check(err)

  w.Write(b.Content)
}
