
package pics

import (
  "fmt"
  "time"
  "strings"
  "mime"
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
    fmt.Println("debug -1: ")
    pl := buildPicList(cx)
    fmt.Println("debug0: ")
    err := tmpl.Execute(w, pl)
    fmt.Println("debug1: ", err)
    util.Check(err)
  } else if strings.HasPrefix(pth, "pics/ref/") {
    name := path.Base(pth)
    ref := name[:len(name)-len(path.Ext(name))]
    m, data, err := cx.ReconstituteFile(ref)
    util.Check(err)

    ext := path.Ext(m.Name)
    w.Header().Set("Content-Type", mime.TypeByExtension(ext))
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

  nBlobs, nPics := 10, 10
  for skip := 0; len(pl) < nPics; skip += nBlobs {
    indReq.SkipN = skip
    blobs, err := cx.IndexBlobs("time", nBlobs, indReq)
    fmt.Println("debug5: len(blobs)=", len(blobs))
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
    fmt.Println("file: ", m.Name)
    if err != nil {
      fmt.Println("not a valid file blob")
      fmt.Println("why not: ", err)
      continue
    } else if ! validImageFile(&m) {
      fmt.Println("not a valid image blob")
      continue
    }
    fmt.Println("FOUND ONE!!!!!!!!!!!!!!")

    pl = append(pl, &pic{FileName: m.Name, Path: "ref/" + b.Ref() + path.Ext(m.Name)})
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
