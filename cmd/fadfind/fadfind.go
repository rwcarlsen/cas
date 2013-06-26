package main

import (
	"flag"
	"fmt"
	"github.com/rwcarlsen/cas/blob"
	"github.com/rwcarlsen/cas/blobserv"
	"github.com/rwcarlsen/cas/mount"
	"github.com/rwcarlsen/cas/query"
	"log"
	"os"
	"strings"
	"time"
)

var max = flag.Int("max", 0, "maximum number of results to retrieve")
var prefix = flag.String("path", "", "mount blobs under specified path")
var showHidden = flag.Bool("hidden", false, "true to include hidden blobs in result")
var showOld = flag.Bool("hist", false, "false to include histories tip")

var cl *blobserv.Client

var lg = log.New(os.Stderr, "fadfind: ", 0)

func main() {
	flag.Parse()
	url := flag.Arg(0)
	tmp := strings.Split(url, "@")
	userPass := strings.Split(tmp[0], ":")
	if len(userPass) != 2 || len(tmp) != 2 {
		lg.Fatalln("Invalid blobserver address")
	}

	cl = &blobserv.Client{
		User: userPass[0],
		Pass: userPass[1],
		Host: tmp[1],
	}

	err := cl.Dial()
	if err != nil {
		lg.Fatalln("Could not connect to blobserver: ", err)
	}

	fmt.Println(url)

	refs := getMatches()
	for _, ref := range refs {
		fmt.Println(ref)
	}
}

func getMatches() []string {
	q := query.New()
	ft := q.NewFilter(filtFn)
	q.SetRoots(ft)

	q.Open()
	defer q.Close()

	batchN := 1000
	timeout := time.After(10 * time.Second)
	for skip, done := 0, false; !done; skip += batchN {
		blobs, err := cl.BlobsBackward(time.Now(), batchN, skip)
		if len(blobs) > 0 {
			q.Process(blobs...)
		}
		if *max > 0 && len(q.Results) == *max {
			break
		}

		if err != nil {
			break
		}
		select {
		case <-timeout:
			done = true
		default:
		}
	}
	if *max > 0 {
		q.Results = q.Results[:*max]
	}
	return blob.RefsFor(q.Results)
}

func filtFn(b *blob.Blob) bool {
	f := blob.NewMeta()
	err := blob.Unmarshal(b, f)
	if err != nil {
		return false
	}

	mm := &mount.Meta{}
	err = f.GetNotes(mount.Key, mm)
	if err != nil {
		mm = &mount.Meta{}
	}

	if !strings.HasPrefix(mm.Path, strings.Trim(*prefix, "./\\")) {
		return false
	} else if !*showOld && !isTip(b) {
		return false
	} else if !*showHidden && mm.Hidden {
		return false
	}
	return true
}

func isTip(b *blob.Blob) bool {
	objref := b.ObjectRef()
	tip, err := cl.ObjectTip(objref)
	if err != nil {
		return false
	}
	return b.Ref() == tip.Ref()
}
