
package blobserver

import (
  "fmt"
  "time"
  "io/ioutil"
  "errors"
  "net/http"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
  "github.com/rwcarlsen/cas/index"
  "github.com/rwcarlsen/cas/auth"
  "github.com/rwcarlsen/cas/util"
)

const (
  defaultDb = "~/.rcas"
  defaultAddr = "0.0.0.0:7777"
  defaultReadTimeout = 60 * time.Second
  defaultWriteTimeout = 60 * time.Second
  defaultHeaderMax = 1 << 20 // 1 Mb
)

const (
  ActionStatus = "Action-Status"
  ActionFailed = "blob get/put failed"
  ActionSuccess = "blob get/put succeeded"
)

const (
  GetField = "Blob-Ref"
)

var (
  DupIndexNameErr = errors.New("blobserver: index name already exists.")
)

type BlobServer struct {
  Db *blobdb.Dbase
  Serv *http.Server
  inds map[string]index.Index
}

func (bs *BlobServer) AddIndex(name string, ind index.Index) error {
  if bs.inds == nil {
    bs.inds = make(map[string]index.Index, 0)
  } else {
    if _, ok := bs.inds[name]; ok {
      return DupIndexNameErr
    }
  }

  bs.inds[name] = ind
  return nil
}

func (bs *BlobServer) notify(blobs ...*blob.Blob) {
  for _, ind := range bs.inds {
    ind.Notify(blobs...)
  }
}

func (bs *BlobServer) ListenAndServe() error {
  if bs.inds == nil {
    bs.inds = make(map[string]index.Index, 0)
  }

  if bs.Db == nil {
    bs.Db, _ = blobdb.New(defaultDb)
  }

  if bs.Serv == nil {
    bs.Serv = &http.Server{
      Addr: defaultAddr,
      ReadTimeout: defaultReadTimeout,
      WriteTimeout: defaultWriteTimeout,
      MaxHeaderBytes: defaultHeaderMax,
    }
  }

  http.Handle("/", &defHandler{})
  http.Handle("/get/", auth.Handler{&getHandler{bs: bs}})
  http.Handle("/put/", auth.Handler{&putHandler{bs: bs}})
  http.Handle("/index/", auth.Handler{&indexHandler{bs: bs}})
  http.Handle("/share/", &shareHandler{bs: bs})

  return bs.Serv.ListenAndServe()
}

type defHandler struct {}

func (h *defHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  fmt.Println(req)
  w.Write([]byte("Page doesn't exist"))
}

type getHandler struct {
  bs *BlobServer
}

func (h *getHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  defer func() {
    if r := recover(); r != nil {
      w.Header().Set(ActionStatus, ActionFailed)
      fmt.Println("blob post failed: ", r)
    }
  }()
  ref := req.FormValue(GetField)

  b, err := h.bs.Db.Get(ref)
  util.Check(err)

  w.Header().Set(ActionStatus, ActionSuccess)
  w.Write(b.Content)
  fmt.Println("successful retrieval")
}

type putHandler struct {
  bs *BlobServer
}

func (h *putHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  defer func() {
    if r := recover(); r != nil {
      w.Header().Set(ActionStatus, ActionFailed)
      fmt.Println("blob post failed: ", r)
    }
  }()

  body, err := ioutil.ReadAll(req.Body)
  util.Check(err)

  b := blob.NewRaw(body)
  err = h.bs.Db.Put(b)
  util.Check(err)
  h.bs.notify(b)

  w.Header().Set(ActionStatus, ActionSuccess)
  fmt.Println("successful post")
}

type indexHandler struct {
  bs *BlobServer
}

func (h *indexHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  defer util.DeferWrite(w)

  //qname := req.FormValue("query")

  // retrieve query results

  //w.Write(results)
}

type shareHandler struct {
  bs *BlobServer
}

func (h *shareHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  defer func() {
    if r := recover(); r != nil {
      fmt.Println(r)
      msg := "blob retrieval failed: " + r.(error).Error()
      m := make(blob.MetaData)
      m["message"] = msg
      resp, _ := blob.Marshal(m)
      w.Write(resp.Content)
    }
  }()

  ref := req.FormValue("ref")
  b, err := h.bs.Db.Get(ref)
  util.Check(err)

  //m := make(blob.MetaData)
  //m, err := blob.Unmarshal(b, &m)
  //util.Check(err)

  //tpe, ok := m[blob.TypeField]
  //if !ok {
  //  // unauthorized
  //  return
  //}

  //if tpe != blob.ShareKind {
  //  // unauthorized
  //  return
  //}

  fname := "what a name"

  head := w.Header()
  head.Set("Content-Type", "application/octet-stream")
  head.Set("Content-Disposition", "attachment; filename=\"" + fname + "\"")
  head.Set("Content-Transfer-Encoding", "binary")
  head.Set("Accept-Ranges", "bytes")
  head.Set("Cache-Control", "private")
  head.Set("Pragma", "private")
  head.Set("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")

  w.Write(b.Content)
}
