
package main

import (
  "os"
  "io"
  "io/ioutil"
  "fmt"
  "net/http"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
)

const (
  dbServer = "localhost"
)

var (
  db, _ = blobdb.New("./dbase")
)

func main() {
  http.HandleFunc("/cas", indexHandler)
  http.HandleFunc("/cas/cas.js", jsHandler)
  http.HandleFunc("/cas/get", get)
  http.HandleFunc("/cas/put", put)
  http.HandleFunc("/cas/putfiles/", putfiles)

  fmt.Println("Starting http server...")
  err := http.ListenAndServe("0.0.0.0:8888", nil)
  if err != nil {
    fmt.Println(err)
    return
  }
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
  f, err := os.Open("index.html")
  if err != nil {
    fmt.Println(err)
    return
  }
  _, err = io.Copy(w, f)
  if err != nil {
    fmt.Println(err)
  }
}

func jsHandler(w http.ResponseWriter, req *http.Request) {
  w.Header().Set("Content-Type", "text/javascript")

  f, err := os.Open("cas.js")
  if err != nil {
    fmt.Println(err)
    return
  }
  _, err = io.Copy(w, f)
  if err != nil {
    fmt.Println(err)
  }
}

func put(w http.ResponseWriter, req *http.Request) {
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    fmt.Println(err)
    w.Write([]byte(err.Error()))
    return
  }

  b := blob.Raw(body)
  err = db.Put(b)
  if err != nil {
    fmt.Println(err)
    w.Write([]byte(err.Error()))
    return
  }

  w.Write([]byte(b.String()))
}

func get(w http.ResponseWriter, req *http.Request) {
  ref, err := ioutil.ReadAll(req.Body)
  if err != nil {
    fmt.Println(err)
    w.Write([]byte(err.Error()))
    return
  }

  b, err := db.Get(string(ref))
  if err != nil {
    fmt.Println(err)
    w.Write([]byte(err.Error()))
    return
  }
  w.Write(b.Content)
}

func putfiles(w http.ResponseWriter, req *http.Request) {
  fmt.Println(req)

  //f, _ := os.Create("./upload/"+header.Filename)
  //defer f.Close()
  //io.Copy(f,fn)

  //req.ParseMultipartForm(10000000)
	mr, err := req.MultipartReader()
  if err != nil {
    fmt.Println(err)
    return
  }
	part, err := mr.NextPart()
	for err == nil {
		if name := part.FormName(); name != "" {
			if part.FileName() != "" {
        data, _ := ioutil.ReadAll(part)
        fmt.Println("filename:", part.FileName())
        fmt.Println(string(data))
				//fileInfos = append(fileInfos, handleUpload(r, part))
			} else {
        //fmt.Println(r.Form)
				//r.Form[name] = append(r.Form[name], getFormValue(part))
			}
		}
		part, err = mr.NextPart()
	}
	return
}
