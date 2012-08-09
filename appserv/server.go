
package appserv

import (
  "fmt"
  "path/filepath"
  "net/http"
  "strings"
  "errors"
  "sort"
  "html/template"
  "io/ioutil"
  "github.com/rwcarlsen/cas/util"
  "github.com/rwcarlsen/cas/blobserv"
)

const tmplDir = "templates"
var defaultClient *blobserv.Client = &blobserv.Client{"https://0.0.0.0:7777", "robert", "password"}

type HandleFunc func(*blobserv.Client, http.ResponseWriter, *http.Request)
var handlers = make(map[string]HandleFunc)
var tmpl *template.Template
var applist []string
var header []byte
var footer []byte
var static string

func Static(path string) string {
  return filepath.Join(static, path)
}

func SetStatic(path string) {
  static = path
}

func RegisterApp(name string, h HandleFunc) error {
  if _, ok := handlers[name]; ok {
    return errors.New("Registration failed: duplicate app name.")
  }
  handlers[name] = h
  return nil
}

func ListenAndServe() error {
  servInit()
  http.HandleFunc("/", handler)

  fmt.Println("Starting http server...")
  err := http.ListenAndServe("0.0.0.0:8888", nil)

  if err != nil {
    return err
  }
  return nil
}

func servInit() {
  tmpl = template.Must(template.ParseFiles(Static("index.tmpl")))
  _ = template.Must(tmpl.ParseGlob(Static(tmplDir + "/*.tmpl")))

  applist = make([]string, 0)
  for name, _ := range handlers {
    applist = append(applist, name)
  }

  sort.Strings(applist)

  var err error
  header, err = ioutil.ReadFile(Static("header.html"))
  util.Check(err)
  footer, err = ioutil.ReadFile(Static("footer.html"))
  util.Check(err)
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
    err := util.LoadStatic(Static(pth), w)
    util.Check(err)
  }
}

