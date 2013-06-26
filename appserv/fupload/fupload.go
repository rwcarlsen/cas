package fupload

import (
	"encoding/json"
	"github.com/rwcarlsen/cas/appserv"
	"github.com/rwcarlsen/cas/blob"
	"github.com/rwcarlsen/cas/blobserv"
	"github.com/rwcarlsen/cas/util"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
)

func Handler(c *blobserv.Client, w http.ResponseWriter, r *http.Request) {
	defer util.DeferWrite(w)

	pth := strings.Trim(r.URL.Path, "/")
	if pth == "fupload" {
		err := util.LoadStatic(appserv.Static("fupload/index.html"), w)
		util.Check(err)
	} else if pth == "fupload/putfiles" {
		putfiles(c, w, r)
	} else {
		err := util.LoadStatic(appserv.Static(pth), w)
		util.Check(err)
	}
}

func putfiles(c *blobserv.Client, w http.ResponseWriter, req *http.Request) {
	defer util.DeferPrint()

	mr, err := req.MultipartReader()
	util.Check(err)

	resps := []interface{}{}

	for part, err := mr.NextPart(); err == nil; {
		if name := part.FormName(); name == "" {
			continue
		} else if part.FileName() == "" {
			continue
		}
		resp := sendFileBlobs(c, part)
		resps = append(resps, resp)
		part, err = mr.NextPart()
	}

	data, _ := json.Marshal(resps)
	w.Write(data)
}

func sendFileBlobs(c *blobserv.Client, part *multipart.Part) (respMeta map[string]interface{}) {
	meta := blob.NewMeta()
	defer func() {
		respMeta = map[string]interface{}{}
		respMeta["name"] = meta.Name
		respMeta["size"] = meta.Size

		if r := recover(); r != nil {
			respMeta["error"] = r.(error).Error()
		}
	}()

	obj := blob.NewObject()
	meta.RcasObjectRef = obj.Ref()
	meta.Name = part.FileName()

	data, err := ioutil.ReadAll(part)
	util.Check(err)

	meta.Size = int64(len(data))

	blobs := blob.SplitRaw(data, blob.DefaultChunkSize)
	meta.ContentRefs = blob.RefsFor(blobs)

	m, err := blob.Marshal(meta)
	util.Check(err)

	blobs = append(blobs, m, obj)
	for _, b := range blobs {
		err = c.PutBlob(b)
		util.Check(err)
	}

	return respMeta
}
