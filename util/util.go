
package util

import (
  "os"
  "net/http"
  "fmt"
  "path"
  "io"
  "mime"
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

func LoadStatic(pth string, w http.ResponseWriter) error {
  f, err := os.Open(pth)
  if err != nil {
    return err
  }

  if ext := path.Ext(pth); ext != "" {
    w.Header().Set("Content-Type", mime.TypeByExtension(ext))
  } else {
    data := make([]byte, 512)
    _, err := f.Read(data)
    Check(err)
    tp := http.DetectContentType(data)
    _, err = f.Seek(0, 0)
    Check(err)
    w.Header().Set("Content-Type", tp)
  }

  _, err = io.Copy(w, f)
  return err
}

