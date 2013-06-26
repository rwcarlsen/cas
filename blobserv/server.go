package blobserv

import (
	"errors"
	"fmt"
	"github.com/rwcarlsen/cas/auth"
	"github.com/rwcarlsen/cas/blobdb"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path"
	"sort"
	"strconv"
	"time"
)

const (
	DefaultDb           = "~/.rcas"
	DefaultAddr         = "0.0.0.0:7777"
	defaultReadTimeout  = 60 * time.Second
	defaultWriteTimeout = 60 * time.Second
	defaultHeaderMax    = 1 << 20 // 1 Mb
)

const (
	ActionStatus  = "Action-Status"
	ActionFailed  = "blob get/put failed"
	ActionSuccess = "blob get/put succeeded"
)

const (
	GetField         = "Blob-Ref"
	ResultCountField = "Num-Index-Results"
	BoundaryField    = "Blob-Boundary"
)

func ListenAndServe(addr string, db blobdb.Interface) error {
	bs := configServ(addr, dbPath)
	return bs.ListenAndServe()
}

func ListenAndServeTLS(addr, db blobdb.Interface, certFile, keyFile string) error {
	bs := configServ(addr, dbPath)
	return bs.ListenAndServeTLS(certFile, keyFile)
}

func configServ(addr, db blobdb.Interface) *Server {
	serv := defaultHttpServer()
	serv.Addr = addr
	return &Server{Db: db, Serv: serv}
}

func defaultHttpServer() *http.Server {
	return &http.Server{
		Addr:           DefaultAddr,
		ReadTimeout:    defaultReadTimeout,
		WriteTimeout:   defaultWriteTimeout,
		MaxHeaderBytes: defaultHeaderMax,
	}
}

type Server struct {
	Db        *blobdb.Interface
	Serv      *http.Server
	listeners []*Client
}

func (bs *Server) AddListener(c *Client) {
	listeners = append(listeners, addr)
}

func (bs *Server) notifyListeners(blob []byte) {
	for _, c := range bs.listeners {
		err := c.SendBlob(blob)
	}
}

func (bs *Server) ListenAndServe() error {
	bs.prepareListen()
	return bs.Serv.ListenAndServe()
}

func (bs *Server) ListenAndServeTLS(certFile, keyFile string) error {
	bs.prepareListen()
	return bs.Serv.ListenAndServeTLS(certFile, keyFile)
}

func (bs *Server) prepareListen() {
	if bs.Serv == nil {
		bs.Serv = defaultHttpServer()
	}

	r := mux.NewRouter()
	r.HandleFunc("/", defHandler)
	r.Handle("/get/{ref:.*}", auth.Handler{&getHandler{bs: bs}})
	r.Handle("/put/", auth.Handler{&putHandler{bs: bs}})
	r.Handle("/enumerate/after:{after:.*}/limit:{limit:[0-9]*}", auth.Handler{&enumHandler{bs: bs}})
	bs.Serv.Handler = r
}

func defHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("rwc cas blobserver"))
}

type getHandler struct {
	bs *Server
}

func (h *getHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(r)
	vars["ref"]

	r, err := h.bs.Db.Get(vars["ref"])
	_, err = io.Copy(w, r)
}

func (h *getHandler) Unauthorized(w http.ResponseWriter, req *http.Request) {
	auth.SendUnauthorized(w)
}

type putHandler struct {
	bs *Server
}

func (h *putHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	ref, n, err := h.bs.Db.Put(bytes.NewBuffer(data))
	h.bs.notifyListeners(data)
}

func (h *putHandler) Unauthorized(w http.ResponseWriter, req *http.Request) {
	auth.SendUnauthorized(w)
}

type enumHandler struct {
	bs *Server
}

func (h *enumHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
}

func (h *enumHandler) Unauthorized(w http.ResponseWriter, req *http.Request) {
	auth.SendUnauthorized(w)
}
