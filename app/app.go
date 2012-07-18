
package app

import (
  "bytes"
  "io/ioutil"
  "net/http"
  "errors"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobserver"
)

type HandleFunc func(*Context, http.ResponseWriter, *http.Request)

type Context struct {
  BlobServerHost string
  User string
  Pass string
}

func (cx *Context) setAuth(r *http.Request) {
  if cx.User + cx.Pass != "" {
    r.SetBasicAuth(cx.User, cx.Pass)
  }
}

func (cx *Context) GetBlobContent(ref string) (content []byte, err error) {
  r, err := http.NewRequest("GET", cx.BlobServerHost, nil)
  if err != nil {
    return nil, err
  }

  r.URL.Path = "/get/"
  r.Header.Set(blobserver.GetField, ref)
  cx.setAuth(r)

  client := &http.Client{}
  resp, err := client.Do(r)
  if err != nil {
    return content, err
  }

  status := resp.Header.Get(blobserver.ActionStatus)
  if status == blobserver.ActionFailed {
    return content, errors.New("app: blob retrieval failed")
  }

  content, err = ioutil.ReadAll(resp.Body)
  if err != nil {
    return content, err
  }

  resp.Body.Close()
  return content, nil
}

func (cx *Context) PutBlob(b *blob.Blob) error {
  r, err := http.NewRequest("POST", cx.BlobServerHost, bytes.NewBuffer(b.Content))
  if err != nil {
    return err
  }

  r.URL.Path = "/put/"
  cx.setAuth(r)
  client := &http.Client{}
  resp, err := client.Do(r)
  if err != nil {
    return err
  }

  status := resp.Header.Get(blobserver.ActionStatus)
  if status == blobserver.ActionFailed {
    return errors.New("app: blob posting failed")
  }

  return nil
}
