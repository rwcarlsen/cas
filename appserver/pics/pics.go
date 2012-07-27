
package pics

import (
  "time"
  "strings"
  "mime"
  "net/http"
  "html/template"
  "path"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/blobserv"
  "github.com/rwcarlsen/cas/timeindex"
  "github.com/rwcarlsen/cas/objindex"
  "github.com/rwcarlsen/cas/appserver/pics/photos"
)

var tmpl *template.Template
func init() {
  tmpl = template.Must(template.ParseFiles("pics/index.tmpl"))
}

var picIndex *photos.Index
var c *blobserv.Client
func Handle(nc *blobserv.Client, w http.ResponseWriter, r *http.Request) {
  defer util.DeferWrite(w)
  c = nc

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
    fref := c.ObjectTip(p.FileObjRef).Ref()

    m, data, err := c.ReconstituteFile(fref)
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
    Time: picIndex.LastUpdate,
    Dir:timeindex.Forward,
  }

  nBlobs := 50
  for skip := 0; true; skip += nBlobs {
    indReq.SkipN = skip
    blobs, err := c.IndexBlobs("time", nBlobs, indReq)
    if err != nil {
      break
    }

    //picIndex.Notify(blobs...)

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
    blobs, err := c.IndexBlobs("time", nBlobs, indReq)
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
  err := c.PutBlob(obj)
  if err != nil {
    panic("pics: could not create photo index")
  }
}

func picForObj(ref string) *photos.Photo {
  b := c.ObjectTip(ref)
  p := photos.NewPhoto()
  err := blob.Unmarshal(b, p)
  util.Check(err)
  return p
}

func picLinks(refs []string) map[string]*photos.Photo {
  links := map[string]*photos.Photo{}
  for _, ref := range refs {
    links["ref/" + ref + ".photo"] = picForObj(ref)
  }
  return links
}

