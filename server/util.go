
package main

import (
  "os"
  "net/http"
  "fmt"
  "path"
)

func deferPrint() {
  if r := recover(); r != nil {
    fmt.Println(r)
  }
}

func deferWrite(w http.ResponseWriter) {
  if r := recover(); r != nil {
    fmt.Println(r)
    w.Write([]byte(r.(error).Error()))
  }
}

func check(err error) {
  if err != nil {
    panic(err)
  }
}

func contentType(pth string, f *os.File) (ctype string) {
  ext := path.Ext(pth)
  if ext == ".js" {
    ctype = "text/javascript"
  } else if ext == ".css" {
    ctype = "text/css"
  } else {
    data := make([]byte, 512)
    _, err := f.Read(data)
    check(err)
    ctype = http.DetectContentType(data)
    _, err = f.Seek(0, 0)
    check(err)
  }
  return
}

