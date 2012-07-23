
package pics

import (
  "time"
  "strings"
  "net/http"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/app"
  "github.com/rwcarlsen/cas/timeindex"
  "html/template"
  "path"
)

var tmpl *template.Template
func init() {
  tmpl = template.Must(template.ParseFiles("pics/index.tmpl"))
}

func Handle(cx *app.Context, w http.ResponseWriter, r *http.Request) {
  defer util.DeferWrite(w)

  pth := strings.Trim(r.URL.Path, "/")
  if pth == "pics" {
    pl := buildPicList(cx)
    err := tmpl.Execute(w, pl)
    util.Check(err)
  } else if strings.HasPrefix(pth, "pics/ref/") {
    ref := path.Base(pth)
    data, err := cx.GetBlobContent(ref)
    util.Check(err)
    w.Write(data)
  } else {
    err := util.LoadStatic(pth, w)
    util.Check(err)
  }
}

type pic struct {
  FileName string
  Path string
}

func buildPicList(cx *app.Context) []*pic {
  pl := []*pic{}

  indReq := timeindex.Request{
    Time: time.Now(),
    Dir:timeindex.Backward,
  }

  nBlobs, nPics := 20, 20
  for len(pl) < nPics {
    blobs, err := cx.IndexBlobs("time", nBlobs, indReq)
    util.Check(err)

    pics := makePics(blobs)
    pl = append(pl, pics...)

    if len(blobs) < nBlobs {
      break
    }
  }

  return pl
}

func makePics(blobs []*blob.Blob) []*pic {
  pl := []*pic{}
  for _, b := range blobs {
    m := blob.FileMeta{}
    err := blob.Unmarshal(b, &m)
    if err != nil {
      continue
    } else if ! validImageFile(&m) {
      continue
    }

    pl = append(pl, &pic{FileName: m.Name, Path: "ref/" + b.Ref()})
  }
  return pl
}

func validImageFile(m *blob.FileMeta) bool {
  if m.RcasType != blob.FileType {
    return false
  }
  switch strings.ToLower(path.Ext(m.Name)) {
    case ".jpg", ".jpeg", ".gif", ".png": return true
  }
  return false
}
