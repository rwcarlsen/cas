
package app

import (
  "bytes"
  "strconv"
  "mime/multipart"
  "encoding/json"
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

func (cx *Context) GetBlob(ref string) (b *blob.Blob, err error) {
  data, err := cx.GetBlobContent(ref)
  if err != nil {
    return nil, err
  }

  return blob.NewRaw(data), nil
}

func (cx *Context) GetBlobContent(ref string) (content []byte, err error) {
  r, err := http.NewRequest("GET", cx.BlobServerHost, nil)
  if err != nil {
    return nil, err
  }

  r.URL.Path = "/ref/" + ref
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

func (cx *Context) ReconstituteFile(ref string) (m *blob.FileMeta, content []byte, err error) {
  b, err := cx.GetBlob(ref)
  if err != nil {
    return nil, nil, err
  }

  m = &blob.FileMeta{}
  err = blob.Unmarshal(b, m)
  if err != nil {
    return nil, nil, err
  }

  chunks := []*blob.Blob{}
  for _, ref := range m.ContentRefs {
    b, err := cx.GetBlob(ref)
    if err != nil {
      return nil, nil, err
    }
    chunks = append(chunks, b)
  }

  return m, blob.Reconstitute(chunks...), nil
}

func (cx *Context) PutBlob(b *blob.Blob) error {
  r, err := http.NewRequest("POST", cx.BlobServerHost, bytes.NewBuffer(b.Content()))
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

func (cx *Context) IndexBlobs(name string, nBlobs int, params interface{}) (blobs []*blob.Blob, err error) {
  data, err := json.Marshal(params)
  if err != nil {
    return nil, err
  }

  r, err := http.NewRequest("POST", cx.BlobServerHost, bytes.NewBuffer(data))
  if err != nil {
    return nil, err
  }

  r.URL.Path = "/index/"
  r.Header.Set(blobserver.IndexField, name)
  r.Header.Set(blobserver.ResultCountField, strconv.Itoa(nBlobs))
  cx.setAuth(r)

  client := &http.Client{}
  resp, err := client.Do(r)
  if err != nil {
    return nil, err
  }

  boundary := resp.Header.Get(blobserver.BoundaryField)
  mr := multipart.NewReader(resp.Body, boundary)

  blobs = []*blob.Blob{}
	for {
		part, err := mr.NextPart()
    if err != nil {
      break
    }

    if part.FileName() == "" {
      continue
    }

    data, err := ioutil.ReadAll(part)
    if err != nil {
      return nil, err
    }

    blobs = append(blobs, blob.NewRaw(data))
	}

  return blobs, nil
}



