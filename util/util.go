
package util

import (
  "os"
  "net/http"
  "fmt"
  "path"
  "io"
)

func DeferPrint() {
  if r := recover(); r != nil {
    fmt.Println(r)
  }
}

func DeferWrite(w http.ResponseWriter) {
  if r := recover(); r != nil {
    fmt.Println(r)
    w.Write([]byte(r.(error).Error()))
  }
}

func Check(err error) {
  if err != nil {
    panic(err)
  }
}

func ContentType(f *os.File) (ctype string) {
  ext := path.Ext(f.Name())
  if ext == ".js" {
    ctype = "text/javascript"
  } else if ext == ".html" {
    ctype = "text/html"
  } else if ext == ".htm" {
    ctype = "text/html"
  } else if ext == ".css" {
    ctype = "text/css"
  } else {
    data := make([]byte, 512)
    _, err := f.Read(data)
    Check(err)
    ctype = http.DetectContentType(data)
    _, err = f.Seek(0, 0)
    Check(err)
  }
  return
}

func LoadStatic(pth string, w http.ResponseWriter) error {
  f, err := os.Open(pth)
  if err != nil {
    return err
  }

  w.Header().Set("Content-Type", ContentType(f))

  _, err = io.Copy(w, f)
  return err
}

