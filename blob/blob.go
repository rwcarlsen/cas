
package blob

import (
  "crypto"
  "crypto/sha256"
  "crypto/sha512"
  "encoding/hex"
  "encoding/json"
  "time"
  "errors"
)

const (
  NameHashSep = "-"
  DefaultHash = crypto.SHA256
)

// universal meta blob fields
const (
  Version = "rcasVersion"
  CurrVersion = "0.1"
  Timestamp = "rcasTimestamp"
  TimeFormat = time.RFC3339
  ObjectRef = "objectRef"
)

// universal TypeField values
const (
  Type = "rcasType"
  File = "file" // generic meta type referring to bytes payload
  MetaNode = "meta-node" // meta type with no bytes payload
  Share = "share" // defines permissions for sharing a target blob
  Object = "object" // random, arbitrary blob used to simulate mutability
)

var (
  hash2Name = map[crypto.Hash]string { }
  name2Hash = map[string]crypto.Hash { }
)

func init() {
  hash2Name[crypto.SHA256] = "sha256"
  hash2Name[crypto.SHA512] = "sha512"
  crypto.RegisterHash(crypto.SHA256, sha256.New)
  crypto.RegisterHash(crypto.SHA512, sha512.New)
  for h, n := range hash2Name {
    name2Hash[n] = h
  }
}

func HashToName(h crypto.Hash) string {
  return hash2Name[h]
}

func NameToHash(n string) crypto.Hash {
  return name2Hash[n]
}

// generic type for creating key-value based blobs
type MetaData map[string] interface{}

// Marshal creates a time-stamped, json encoded blob from v.
func Marshal(v interface{}) (b *Blob, err error) {
  data, err := json.Marshal(v)
  if err != nil {
    return nil, err
  }

  b = NewRaw(data)

  err = b.set(Timestamp, time.Now().Format(TimeFormat))
  if err != nil {
    return nil, err
  }

  err = b.set(Version, CurrVersion)
  if err != nil {
    return nil, err
  }

  return b, nil
}

// Unmarshal parses a json encoded blob and stores the result in 
// the value pointed to by v.
func Unmarshal(b *Blob, v interface{}) error {
  return json.Unmarshal(b.Content, v)
}

type Blob struct {
  Hash crypto.Hash
  Content []byte
}

// Raw creates a blob using the DefaultHash holding the passed content.
func NewRaw(content []byte) *Blob {
  return &Blob{Hash: DefaultHash, Content: content}
}

func (b *Blob) Get(prop string) interface{} {
  m := MetaData{}
  err := Unmarshal(b, &m)
  if err != nil {
    return nil
  }

  val, ok := m[prop]
  if !ok {
    return nil
  }

  return val
}

func (b *Blob) set(prop string, val interface{}) error {
  m := MetaData{}
  err := Unmarshal(b, &m)
  if err != nil {
    return err
  }

  if _, ok := m[prop]; ok {
    return errors.New("blob: cannot overwrite existing property")
  }
  m[prop] = val

  return nil
}

// Sum returns the hash sum of the blob's content using its hash function
func (b *Blob) Sum() []byte {
  hsh := b.Hash.New()
  hsh.Write(b.Content)
  return hsh.Sum([]byte{})
}

// Ref returns hash-name + hash for the blob.
func (b *Blob) Ref() string {
  return HashToName(b.Hash) + NameHashSep + hex.EncodeToString(b.Sum())
}

func (b *Blob) String() string {
  return b.Ref() + ":\n" +  string(b.Content)
}

// RefsFor returns a list of the refs for each blob in the given list.
func RefsFor(blobs []*Blob) []string {
  refs := make([]string, len(blobs))
  for i, b := range blobs {
    refs[i] = b.Ref()
  }
  return refs
}

