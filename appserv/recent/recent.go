package recent

import (
	"bytes"
	"encoding/json"
	"github.com/rwcarlsen/cas/appserv"
	"github.com/rwcarlsen/cas/blobserv"
	"github.com/rwcarlsen/cas/util"
	"html/template"
	"net/http"
	"strings"
	"time"
)

func Handler(c *blobserv.Client, w http.ResponseWriter, r *http.Request) {
	defer util.DeferWrite(w)

	tmpl := template.Must(template.ParseFiles(appserv.Static("recent/index.tmpl")))

	pth := strings.Trim(r.URL.Path, "/")
	if pth == "recent" {
		data := stripBlobs(c)
		err := tmpl.Execute(w, data)
		util.Check(err)
	} else {
		err := util.LoadStatic(appserv.Static(pth), w)
		util.Check(err)
	}
}

type shortblob struct {
	Ref     string
	Content string
}

func stripBlobs(c *blobserv.Client) []*shortblob {
	blobs, err := c.BlobsBackward(time.Now(), 20, 0)
	util.Check(err)

	short := []*shortblob{}
	for _, b := range blobs {
		buf := bytes.NewBuffer([]byte{})
		json.Indent(buf, b.Content(), "", "    ")
		short = append(short, &shortblob{b.Ref(), buf.String()})
	}

	return short
}
