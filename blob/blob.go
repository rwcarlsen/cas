
package blob

import (
  "crypto"
  "crypto/sha256"
  "crypto/sha512"
  "encoding/hex"
  "encoding/json"
  "time"
)

const (
  NameHashSep = "-"
  DefaultHash = crypto.SHA256
)

// universal meta blob fields
const (
  VersionField = "rcasVersion"
  Version = "0.1"
  TimeField = "rcasTimestamp"
  TimeFormat = time.RFC3339Nano
)

type Type string

// universal TypeField values
const (
  FileType Type = "file"
  NoteType = "note"
  NoneType = "none"
  ShareType = "share"
  ObjectType = "object"
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

  var m MetaData
  err = json.Unmarshal(data, &m)
  if err != nil {
    return nil, err
  }

  m[TimeField] = time.Now().UTC().Format(TimeFormat)
  m[VersionField] = Version

  data, err = json.Marshal(m)
  if err != nil {
    return nil, err
  }

  return NewRaw(data), nil
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

