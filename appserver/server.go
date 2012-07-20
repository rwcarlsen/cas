
package main

import (
  "fmt"
  "net/http"
  "strings"
  "sort"
  "html/template"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/app"
  "github.com/rwcarlsen/cas/appserver/notedrop"
  "github.com/rwcarlsen/cas/appserver/fupload"
)

var defaultContext *app.Context = &app.Context{"http://rwc-server.dyndns.org:7777", "robert", "password"}

var handlers map[string]app.HandleFunc

// add new apps by listing them here in this init func
func init() {
  handlers = map[string]app.HandleFunc{}
  handlers["notedrop"] = notedrop.Handle
  handlers["fupload"] = fupload.Handle
}

var tmpl *template.Template
var applist []string
func init() {
  tmpl = template.Must(template.ParseFiles("index.tmpl", "applinks.tmpl"))

  applist = make([]string, 0)
  for name, _ := range handlers {
    applist = append(applist, name)
  }

  sort.Strings(applist)
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
    err := tmpl.Execute(w, applist)
    util.Check(err)
  } else if _, ok := handlers[base]; ok {
    handlers[base](defaultContext, w, r)
  } else {
    err := util.LoadStatic(pth, w)
    util.Check(err)
  }
}

