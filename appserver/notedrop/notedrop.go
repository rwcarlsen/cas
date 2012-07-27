
package notedrop

import (
  "strings"
  "io/ioutil"
  "net/http"
  "encoding/json"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/blobserv"
)

const myType = "note-drop"

func Handle(c *blobserv.Client, w http.ResponseWriter, r *http.Request) {
  defer util.DeferWrite(w)

  pth := strings.Trim(r.URL.Path, "/")
  if pth == "notedrop" {
    err := util.LoadStatic("notedrop/index.html", w)
    util.Check(err)
  } else if pth == "notedrop/putnote" {
    putnote(c, w, r)
  } else {
    err := util.LoadStatic(pth, w)
    util.Check(err)
  }
}

func putnote(c *blobserv.Client, w http.ResponseWriter, req *http.Request) {
  defer util.DeferWrite(w)

  body, err := ioutil.ReadAll(req.Body)
  util.Check(err)

  var note map[string]interface{}
  err = json.Unmarshal(body, &note)
  util.Check(err)

  note[blob.Type] = myType

  b, err := blob.Marshal(note)
  util.Check(err)

  err = c.PutBlob(b)
  util.Check(err)

  w.Write(b.Content())
}
