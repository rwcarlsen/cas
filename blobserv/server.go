
package blobserv

import (
  "fmt"
  "sort"
  "time"
  "strconv"
  "errors"
  "path"
  "io/ioutil"
  "net/http"
  "mime/multipart"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
  "github.com/rwcarlsen/cas/blobserv/index"
  "github.com/rwcarlsen/cas/auth"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/blobserv/timeindex"
  "github.com/rwcarlsen/cas/blobserv/objindex"
)

const (
  DefaultDb = "~/.rcas"
  DefaultAddr = "0.0.0.0:7777"
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
  IndexField = "Index-Name"
  ResultCountField = "Num-Index-Results"
  BoundaryField = "Blob-Boundary"
)

var (
  DupIndexNameErr = errors.New("blobserv: index name already exists")
)

func ListenAndServe(addr string, dbPath string) error {
  bs := configServ(addr, dbPath)
  return bs.ListenAndServe()
}

func ListenAndServeTLS(addr, dbPath string, certFile, keyFile string) error {
  bs := configServ(addr, dbPath)
  return bs.ListenAndServeTLS(certFile, keyFile)
}

func configServ(addr, dbPath string) *Server {
  db, _ := blobdb.New(dbPath)
  tInd := timeindex.New()
  oInd := objindex.New()
  for b := range db.Walk() {
    tInd.Notify(b)
    oInd.Notify(b)
  }
  sort.Sort(tInd)
  oInd.Sort()

  serv := defaultHttpServer()
  serv.Addr = addr
  bs := &Server{Db: db, Serv: serv}
  bs.AddIndex("time", tInd)
  bs.AddIndex("object", oInd)
  return bs
}

func defaultHttpServer() *http.Server {
  return &http.Server{
    Addr: DefaultAddr,
    ReadTimeout: defaultReadTimeout,
    WriteTimeout: defaultWriteTimeout,
    MaxHeaderBytes: defaultHeaderMax,
  }
}

type Server struct {
  Db *blobdb.Dbase
  Serv *http.Server
  inds map[string]index.Index
}

func (bs *Server) AddIndex(name string, ind index.Index) error {
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

func (bs *Server) notify(blobs ...*blob.Blob) {
  for _, ind := range bs.inds {
    ind.Notify(blobs...)
  }
}

func (bs *Server) prepareListen() {
  if bs.inds == nil {
    bs.inds = make(map[string]index.Index, 0)
  }

  if bs.Db == nil {
    bs.Db, _ = blobdb.New(DefaultDb)
  }

  if bs.Serv == nil {
    bs.Serv = defaultHttpServer()
  }

  http.Handle("/", &defHandler{})
  http.Handle("/ref/", auth.Handler{&getHandler{bs: bs}})
  http.Handle("/put/", auth.Handler{&putHandler{bs: bs}})
  http.Handle("/index/", auth.Handler{&indexHandler{bs: bs}})
}

func (bs *Server) ListenAndServe() error {
  bs.prepareListen()
  return bs.Serv.ListenAndServe()
}

func (bs *Server) ListenAndServeTLS(certFile, keyFile string) error {
  bs.prepareListen()
  return bs.Serv.ListenAndServeTLS(certFile, keyFile)
}

type defHandler struct {}

func (h *defHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  w.Write([]byte("Page doesn't exist"))
}

type getHandler struct {
  bs *Server
}

func (h *getHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  defer func() {
    if r := recover(); r != nil {
      w.Header().Set(ActionStatus, ActionFailed)
      fmt.Println("blob post failed: ", r)
    }
  }()

  ref := path.Base(req.URL.Path)
  b, err := h.bs.Db.Get(ref)
  util.Check(err)

  w.Header().Set(ActionStatus, ActionSuccess)
  w.Write(b.Content())
  fmt.Println("successful retrieval")
}

// Unauthorized checks for and handles cases where authentication can occur via
// share blobs.
func (h *getHandler) Unauthorized(w http.ResponseWriter, req *http.Request) {
  defer func() {
    if r := recover(); r != nil {
      w.Header().Set(ActionStatus, ActionFailed)
      fmt.Println("blob post failed: ", r)
    }
  }()

  shareRef := req.FormValue("via")
  if shareRef == "" {
    auth.SendUnauthorized(w)
    return
  }

  b, err := h.bs.Db.Get(shareRef)
  if err != nil {
    auth.SendUnauthorized(w)
    return
  }

  share := &blob.Share{}
  err = blob.Unmarshal(b, share)
  util.Check(err)

  ref := path.Base(req.URL.Path)
  b, err = h.bs.Db.Get(ref)
  util.Check(err)

  if !share.AuthorizedGet(b) {
    auth.SendUnauthorized(w)
    return
  }

  w.Header().Set(ActionStatus, ActionSuccess)
  w.Write(b.Content())
  fmt.Println("successful retrieval")
}

type putHandler struct {
  bs *Server
}

func (h *putHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  defer func() {
    if r := recover(); r != nil {
      w.Header().Set(ActionStatus, ActionFailed)
      fmt.Println("blob post issues: ", r)
    }
  }()

  body, err := ioutil.ReadAll(req.Body)
  util.Check(err)

  b := blob.NewRaw(body)
  err = h.bs.Db.Put(b)
  if err == blobdb.HashCollideErr {
    w.Header().Set(ActionStatus, ActionFailed)
    return
  }
  h.bs.notify(b)

  w.Header().Set(ActionStatus, ActionSuccess)
  fmt.Println("successful post")
}

// Unauthorized checks for and handles cases where authentication can occur via
// share blobs.
func (h *putHandler) Unauthorized(w http.ResponseWriter, req *http.Request) {
  auth.SendUnauthorized(w)
}

type indexHandler struct {
  bs *Server
}

func (h *indexHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  defer func() {
    if r := recover(); r != nil {
      w.Header().Set(ActionStatus, ActionFailed)
      fmt.Println("index access failed: ", r)
    }
  }()

  name := req.Header.Get(IndexField)
  ind, ok := h.bs.inds[name]
  if !ok {
    panic("invalid index name")
  }

  it, err := ind.GetIter(req)
  util.Check(err)

  n, err := strconv.Atoi(req.Header.Get(ResultCountField))
  util.Check(err)

  refs := multipart.NewWriter(w)
  w.Header().Set("Content-Type", "multipart/mixed")
  w.Header().Set(BoundaryField, refs.Boundary() )

  defer refs.Close()
  for i := 0; i < n; i++ {
    ref, err := it.Next()
    if err != nil {
      break
    }

    b, err := h.bs.Db.Get(ref)
    util.Check(err)

    part, err := refs.CreateFormFile("blob-ref", ref)
    util.Check(err)
    part.Write(b.Content())
  }
}

func (h *indexHandler) Unauthorized(w http.ResponseWriter, req *http.Request) {
  auth.SendUnauthorized(w)
}

