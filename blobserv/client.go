package blobserv

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	Host string
	User string
	Pass string
}

func (c *Client) setAuth(r *http.Request) {
	if c.User+c.Pass != "" {
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

func getClient() *http.Client {
	// allows me to use encryption certificate that is not signed by a
	// verified authority
	config := &tls.Config{InsecureSkipVerify: true}
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: config, Proxy: http.ProxyFromEnvironment},
	}
}

func (c *Client) GetBlob(ref string) ([]byte, error) {
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

func (c *Client) SendBlob(blobs ...[]byte) error {
	r, err := http.NewRequest("POST", c.Host, bytes.NewBuffer(data))
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
