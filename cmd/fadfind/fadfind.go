
package main

import (
  "fmt"
  "time"
  "flag"
  "strings"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobserv"
  "github.com/rwcarlsen/cas/query"
)

var max = flag.Int("max", 0, "maximum number of results to retrieve")
var prefix = flag.String("path", "", "mount blobs under specified path")

var cl *blobserv.Client

func main() {
  flag.Parse()
  url := flag.Arg(0)
  tmp := strings.Split(url, "@")
  userPass := strings.Split(tmp[0], ":")

  cl = &blobserv.Client{
    User: userPass[0],
    Pass: userPass[1],
    Host: tmp[1],
  }

  err := cl.Dial()
  if err != nil {
    fmt.Println("Could not connect to blobserver: ", err)
    return
  }

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
    if *max > 0  && len(q.Results) == *max {
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
  f := &blob.FileMeta{}
  err := blob.Unmarshal(b, f)
  if err != nil {
    return false
  }

  if !strings.HasPrefix(f.Path, strings.Trim(*prefix, "./\\")) {
    return false
  } else if !isTip(b) {
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

