
package util

import (
  "os"
  "io/ioutil"
  "bytes"
  "net/http"
  "strings"
  "fmt"
  "path"
  "io"
  "mime"
  "log"
)

var Dbg = log.New(os.Stderr, "debug: ", log.Lshortfile | log.Ltime)

func DeferPrint() {
  if r := recover(); r != nil {
    fmt.Println(r)
  }
}

func DeferWrite(w http.ResponseWriter) {
  if r := recover(); r != nil {
    fmt.Println(r)
    w.Write([]byte(fmt.Sprint(r)))
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

func PipedStdin() []string {
  var line string
  var err error
  refs := []string{}
  data, _ := ioutil.ReadAll(os.Stdin)
  buff := bytes.NewBuffer(data)
  for err == nil {
    line, err = buff.ReadString('\n')
    line = strings.TrimSpace(line)
    if len(line) > 0 {
      refs = append(refs, line)
    }
  }
  return refs
}

