
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
  "github.com/rwcarlsen/cas/appserv/pics/photos"
  "github.com/rwcarlsen/cas/appserv"
)

var picIndex *photos.Index
var c *blobserv.Client
func Handler(nc *blobserv.Client, w http.ResponseWriter, r *http.Request) {
  defer util.DeferWrite(w)
  c = nc

  tmpl := template.Must(template.ParseFiles(appserv.Static("pics/index.tmpl")))

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

    fblob, err := c.ObjectTip(ref)
    util.Check(err)

    m, data, err := c.ReconstituteFile(fblob.Ref())
    util.Check(err)

    ext := path.Ext(m.Name)
    w.Header().Set("Content-Type", mime.TypeByExtension(ext))
    w.Write(data)
  } else {
    err := util.LoadStatic(appserv.Static(pth), w)
    util.Check(err)
  }
}

func updateIndex() {
  nBlobs := 50
  for skip := 0; true; skip += nBlobs {
    blobs, err := c.BlobsForward(picIndex.LastUpdate, nBlobs, skip)
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
  nBlobs := 10
  for skip := 0; true; skip += nBlobs {
    blobs, err := c.BlobsBackward(time.Now(), nBlobs, skip)
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
  b, err := c.ObjectTip(ref)
  util.Check(err)
  p := &photos.Photo{}
  err = blob.Unmarshal(b, p)
  util.Check(err)
  return p
}

func picLinks(refs []string) map[string]*photos.Photo {
  links := map[string]*photos.Photo{}
  for _, ref := range refs {
    // fix this to not be a blank photo TODO
    links["ref/" + ref + ".photo"] = &photos.Photo{}
  }
  return links
}

