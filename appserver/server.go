
package main

import (
  "fmt"
  "net/http"
  "strings"
  "sort"
  "html/template"
  "io/ioutil"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/blobserv"
  "github.com/rwcarlsen/cas/appserver/notedrop"
  "github.com/rwcarlsen/cas/appserver/fupload"
  "github.com/rwcarlsen/cas/appserver/recent"
  "github.com/rwcarlsen/cas/appserver/pics"
)

const tmplDir = "templates"
var defaultClient *blobserv.Client = &blobserv.Client{"https://0.0.0.0:7777", "robert", "password"}

type HandleFunc func(*blobserv.Client, http.ResponseWriter, *http.Request)

//// add new apps by listing them here in this init func
func init() {
  handlers = map[string]HandleFunc{}
  handlers["notedrop"] = notedrop.Handle
  handlers["fupload"] = fupload.Handle
  handlers["recent"] = recent.Handle
  handlers["pics"] = pics.Handle
}

var handlers map[string]HandleFunc
var tmpl *template.Template
var applist []string
func init() {
  tmpl = template.Must(template.ParseFiles("index.tmpl"))
  _ = template.Must(tmpl.ParseGlob(tmplDir + "/*.tmpl"))

  applist = make([]string, 0)
  for name, _ := range handlers {
    applist = append(applist, name)
  }

  sort.Strings(applist)
}
var header []byte
var footer []byte
func init() {
  var err error
  header, err = ioutil.ReadFile("header.html")
  util.Check(err)
  footer, err = ioutil.ReadFile("footer.html")
  util.Check(err)
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
    notAjax := r.Header.Get("X-Requested-With") == ""
    notStatic := !strings.Contains(r.URL.Path, ".")
    if notAjax && notStatic {
      w.Write(header)
    }
    handlers[base](defaultClient, w, r)
    if notAjax && notStatic {
      w.Write(footer)
    }
  } else {
    err := util.LoadStatic(pth, w)
    util.Check(err)
  }
}

