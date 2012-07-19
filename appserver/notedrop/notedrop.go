
package notedrop

import (
  "strings"
  "io/ioutil"
  "net/http"
  "encoding/json"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/app"
)

func Handle(cx *app.Context, w http.ResponseWriter, r *http.Request) {
  defer util.DeferWrite(w)

  pth := strings.Trim(r.URL.Path, "/")
  if pth == "notedrop" {
    err := util.LoadStatic("notedrop/index.html", w)
    util.Check(err)
  } else if pth == "notedrop/putnote" {
    putnote(cx, w, r)
  } else {
    err := util.LoadStatic(pth, w)
    util.Check(err)
  }
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
