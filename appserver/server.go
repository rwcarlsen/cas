
package main

import (
  "fmt"
  "net/http"
  "strings"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/app"
  "github.com/rwcarlsen/cas/appserver/notedrop"
  "github.com/rwcarlsen/cas/appserver/fupload"
)

var defaultContext *app.Context = &app.Context{"http://rwc-server.dyndns.org:7777", "robert", "password"}

var handlers map[string]app.HandleFunc

func init() {
  handlers = make(map[string]app.HandleFunc)
  handlers["notedrop"] = notedrop.Handle
  handlers["fupload"] = fupload.Handle
}

func main() {
  http.HandleFunc("/", handler)

  fmt.Println("Starting http server...")
  err := http.ListenAndServe("0.0.0.0:8888", nil)

  if err != nil {
    fmt.Println(err)
    return
  }
}

func handler(w http.ResponseWriter, r *http.Request) {
  defer util.DeferWrite(w)

  pth := strings.Trim(r.URL.Path, "/")
  base := strings.Split(pth, "/")[0]

  if base == "" {
    err := util.LoadStatic("index.html", w)
    util.Check(err)
  } else if _, ok := handlers[base]; ok {
    handlers[base](defaultContext, w, r)
  } else {
    fmt.Println("loading static file", pth)
    err := util.LoadStatic(pth, w)
    util.Check(err)
  }
}

