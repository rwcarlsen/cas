
package blobserv

import (
  "bytes"
  "time"
  "strconv"
  "mime/multipart"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "errors"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobserv/timeindex"
  "github.com/rwcarlsen/cas/blobserv/objindex"
)

type Client struct {
  Host string
  User string
  Pass string
}

func (c *Client) setAuth(r *http.Request) {
  if c.User + c.Pass != "" {
    r.SetBasicAuth(c.User, c.Pass)
  }
}

func (c *Client) GetBlob(ref string) (b *blob.Blob, err error) {
  data, err := c.GetBlobContent(ref)
  if err != nil {
    return nil, err
  }

  return blob.NewRaw(data), nil
}

func (c *Client) GetBlobContent(ref string) (content []byte, err error) {
  r, err := http.NewRequest("GET", c.Host, nil)
  if err != nil {
    return nil, err
  }

  r.URL.Path = "/ref/" + ref
  c.setAuth(r)

  client := &http.Client{}
  resp, err := client.Do(r)
  if err != nil {
    return content, err
  }

  status := resp.Header.Get(ActionStatus)
  if status == ActionFailed {
    return content, errors.New("app: blob retrieval failed")
  }

  content, err = ioutil.ReadAll(resp.Body)
  if err != nil {
    return content, err
  }

  resp.Body.Close()
  return content, nil
}

func (c *Client) ReconstituteFile(ref string) (m *blob.FileMeta, content []byte, err error) {
  b, err := c.GetBlob(ref)
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
    b, err := c.GetBlob(ref)
    if err != nil {
      return nil, nil, err
    }
    chunks = append(chunks, b)
  }

  return m, blob.Reconstitute(chunks...), nil
}

func (c *Client) PutBlob(b *blob.Blob) error {
  r, err := http.NewRequest("POST", c.Host, bytes.NewBuffer(b.Content()))
  if err != nil {
    return err
  }

  r.URL.Path = "/put/"
  c.setAuth(r)
  client := &http.Client{}
  resp, err := client.Do(r)
  if err != nil {
    return err
  }

  status := resp.Header.Get(ActionStatus)
  if status == ActionFailed {
    return errors.New("app: blob posting failed")
  }

  return nil
}

func (c *Client) IndexBlobs(name string, nBlobs int, params interface{}) (blobs []*blob.Blob, err error) {
  data, err := json.Marshal(params)
  if err != nil {
    return nil, err
  }

  r, err := http.NewRequest("POST", c.Host, bytes.NewBuffer(data))
  if err != nil {
    return nil, err
  }

  r.URL.Path = "/index/"
  r.Header.Set(IndexField, name)
  r.Header.Set(ResultCountField, strconv.Itoa(nBlobs))
  c.setAuth(r)

  client := &http.Client{}
  resp, err := client.Do(r)
  if err != nil {
    return nil, err
  }

  boundary := resp.Header.Get(BoundaryField)
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

  if len(blobs) == 0 {
    return nil, errors.New("app: no blobs for that index query")
  }

  return blobs, nil
}

func (c *Client) BlobsBackward(t time.Time, n, nskip int) (b []*blob.Blob, err error) {
  indReq := timeindex.Request{
    Time: t,
    Dir:timeindex.Backward,
    SkipN: nskip,
  }
  return c.IndexBlobs("time", n, &indReq)
}

func (c *Client) BlobsForward(t time.Time, n, nskip int) (b []*blob.Blob, err error) {
  indReq := timeindex.Request{
    Time: t,
    Dir:timeindex.Forward,
    SkipN: nskip,
  }
  return c.IndexBlobs("time", n, &indReq)
}

func (c *Client) ObjectTip(objref string) (b *blob.Blob, err error) {
  objReq := objindex.Request{ObjectRef:objref}
  blobs, err := c.IndexBlobs("object", 1, objReq)
  if err != nil {
    return nil, err
  }
  return blobs[0], nil
}

