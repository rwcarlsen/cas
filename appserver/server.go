
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

  pth := r.URL.Path
  if pth == "/" {
    util.LoadStatic("index.html", w)
  } else {
    name := strings.Split(pth, "/")[1]
    handlers[name](defaultContext, w, r)
  }
}

