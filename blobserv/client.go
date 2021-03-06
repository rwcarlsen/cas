
package blobserv

import (
  "bytes"
  "time"
  "strconv"
  "mime/multipart"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "crypto/tls"
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

func (c *Client) GetBlob(ref string) (*blob.Blob, error) {
  data, err := c.GetBlobContent(ref)
  if err != nil {
    return nil, err
  }
  return blob.NewRaw(data), nil
}

func (c *Client) Dial() error {
  r, err := http.NewRequest("GET", c.Host, nil)
  if err != nil {
    return err
  }
  r.URL.Path = "/ref/foo"
  _, err = getClient().Do(r)
  return err
}

func getClient() *http.Client {
  // allows me to use encryption certificate that is not signed by a
  // verified authority
  config := &tls.Config{InsecureSkipVerify: true}
  return &http.Client{
    Transport: &http.Transport{TLSClientConfig: config, Proxy: http.ProxyFromEnvironment},
  }
}

func (c *Client) GetBlobContent(ref string) ([]byte, error) {
  r, err := http.NewRequest("GET", c.Host, nil)
  if err != nil {
    return nil, err
  }

  r.URL.Path = "/ref/" + ref
  c.setAuth(r)

  resp, err := getClient().Do(r)
  if err != nil {
    return nil, err
  }

  status := resp.Header.Get(ActionStatus)
  if status == ActionFailed {
    return nil, errors.New("blobserv: blob retrieval failed")
  }

  content, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }

  resp.Body.Close()
  return content, nil
}

func (c *Client) ReconstituteFile(ref string) (m *blob.Meta, content []byte, err error) {
  b, err := c.GetBlob(ref)
  if err != nil {
    return nil, nil, err
  }

  m = &blob.Meta{}
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
  resp, err := getClient().Do(r)
  if err != nil {
    return err
  }

  status := resp.Header.Get(ActionStatus)
  if status == ActionFailed {
    return errors.New("blobserv: blob posting failed")
  }

  return nil
}

func (c *Client) IndexBlobs(name string, nBlobs int, params interface{}) ([]*blob.Blob, error) {
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

  resp, err := getClient().Do(r)
  if err != nil {
    return nil, err
  }

  boundary := resp.Header.Get(BoundaryField)
  mr := multipart.NewReader(resp.Body, boundary)

  blobs := []*blob.Blob{}
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
    return nil, errors.New("blobserv: no blobs for that index query")
  }

  return blobs, nil
}

func (c *Client) BlobsBackward(t time.Time, n, nskip int) ([]*blob.Blob, error) {
  indReq := timeindex.Request{
    Time: t,
    Dir:timeindex.Backward,
    SkipN: nskip,
  }
  return c.IndexBlobs("time", n, &indReq)
}

func (c *Client) BlobsForward(t time.Time, n, nskip int) ([]*blob.Blob, error) {
  indReq := timeindex.Request{
    Time: t,
    Dir:timeindex.Forward,
    SkipN: nskip,
  }
  return c.IndexBlobs("time", n, &indReq)
}

func (c *Client) ObjectTip(objref string) (*blob.Blob, error) {
  objReq := objindex.Request{ObjectRef:objref}
  blobs, err := c.IndexBlobs("object", 1, objReq)
  if err != nil {
    return nil, err
  }
  return blobs[0], nil
}

