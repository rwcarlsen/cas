
package pics

import (
  "time"
  "strings"
  "mime"
  "net/http"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/app"
  "github.com/rwcarlsen/cas/timeindex"
  "github.com/rwcarlsen/cas/objindex"
  "github.com/rwcarlsen/cas/appserver/pics/photos"
  "html/template"
  "path"
)

var tmpl *template.Template
func init() {
  tmpl = template.Must(template.ParseFiles("pics/index.tmpl"))
}

var picIndex *photos.Index
var cx *app.Context
func Handle(ncx *app.Context, w http.ResponseWriter, r *http.Request) {
  defer util.DeferWrite(w)
  cx = ncx

  if picIndex == nil {
    loadPicIndex()
  }
  updateIndex()

  pth := strings.Trim(r.URL.Path, "/")
  if pth == "pics" {
    links := picLinks(picIndex.Newest(10))
    err := tmpl.Execute(w, links)
    util.Check(err)
  } else if strings.HasPrefix(pth, "pics/ref/") {
    ref := path.Base(pth)
    ref = ref[:len(ref)-len(path.Ext(ref))]

    p := picForObj(ref)
    fref := tip(p.FileObjRef).Ref()

    m, data, err := cx.ReconstituteFile(fref)
    util.Check(err)

    ext := path.Ext(m.Name)
    w.Header().Set("Content-Type", mime.TypeByExtension(ext))
    w.Write(data)
  } else {
    err := util.LoadStatic(pth, w)
    util.Check(err)
  }
}

func updateIndex() {
  indReq := timeindex.Request{
    Time: picIndex.LastUpdate
    Dir:timeindex.Forward,
  }

  nBlobs := 50
  for skip := 0; true; skip += nBlobs {
    indReq.SkipN = skip
    blobs, err := cx.IndexBlobs("time", nBlobs, indReq)
    if err != nil {
      break
    }

    picIndex.Notify(blobs...)

    if len(blobs) < nBlobs {
      break
    }
  }
  
}

func loadPicIndex() {
  indReq := timeindex.Request{
    Time: time.Now(),
    Dir:timeindex.Backward,
  }

  nBlobs := 10
  for skip := 0; true; skip += nBlobs {
    indReq.SkipN = skip
    blobs, err := cx.IndexBlobs("time", nBlobs, indReq)
    if err != nil {
      break
    }
    for _, b := range blobs {
      if b.Type() == photos.IndexType {
        err := blob.Unmarshal(b, picIndex)
        util.Check(err)
        return
      }
    }

    if len(blobs) < nBlobs {
      break
    }
  }

  // no pre-existing photo index found
  picIndex = photos.NewIndex()
  obj := blob.NewObject()
  picIndex.RcasObjectRef = obj.Ref()
  err := cx.PutBlob(obj)
  if err != nil {
    panic("pics: could not create photo index")
  }
}

func picForObj(ref string) *photos.Photo {
  b := tip(ref)
  p := photos.NewPhoto()
  err := blob.Unmarshal(b, p)
  util.Check(err)
  return p
}

func tip(objref string) *blob.Blob {
  objReq := objindex.Request{ObjectRef:objref}
  blobs, err := cx.IndexBlobs("object", 1, objReq)
  util.Check(err)
  return blobs[0]
}

func picLinks(refs []string) map[string]*photos.Photo {
  links := map[string]*photos.Photo{}
  for _, ref := range refs {
    links["ref/" + ref + ".photo"] = picForObj(ref)
  }
  return links
}

func validImageFile(m *blob.FileMeta) bool {
  if m.RcasType != blob.File {
    return false
  }
  switch strings.ToLower(path.Ext(m.Name)) {
    case ".jpg", ".jpeg", ".gif", ".png": return true
  }
  return false
}
