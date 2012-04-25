
package blobdb

import (
  "../blob"
  "fmt"
  "path"
  "io/ioutil"
)

type dbase struct {
  path string
}

func (db *dbase) Retrieve(id string) *Blob {
  p := path.Join(db.path, id)
  f, err := os.Open(p)

  if err != nil {
    fmt.Println(err)
    return nil
  }

  data := ioutil.ReadAll(f)
  blob := blob.New()
  blob.Write(data)

  return 

}

