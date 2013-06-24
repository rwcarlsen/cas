package blobdb

import (
	"crypto"
	_ "crypto/sha256"
	"fmt"
	"strings"
	"io"
)

var (
	hash2Name = map[crypto.Hash]string{}
	name2Hash = map[string]crypto.Hash{}
)

const (
	DefaultHash = crypto.SHA256
	NameHashSep = "-"
)

func init() {
	hash2Name[crypto.SHA256] = "sha256"
	for h, n := range hash2Name {
		name2Hash[n] = h
	}
}

func hashToName(h crypto.Hash) string {
	return hash2Name[h]
}

func nameToHash(n string) crypto.Hash {
	return name2Hash[n]
}

type Interface interface {
	Get(string) (io.ReadCloser, error)
	Put(r io.Reader) (string, error)
	Enumerate(after string, limit int) []string
}

func MakeBlobRef(r io.Reader) string {
	h := DefaultHash.New()
	if _, err := io.Copy(h, r); err != nil {
		panic(err)
	}
	return hashToName(DefaultHash) + NameHashSep + fmt.Sprintf("%x", h.Sum(nil))
}

func blobRefParts(ref string) (hash crypto.Hash, sum string) {
	parts := strings.Split(ref, NameHashSep)
	if len(parts) != 2 {
		panic("blobdb: Invalid blob ref " + ref)
	}
	return nameToHash(parts[0]), parts[1]
}

func VerifyBlob(data []byte, ref string) bool {
	hash, sum := blobRefParts(ref)
	h := hash.New()
	h.Write(data)

	if fmt.Sprintf("%x", h.Sum(nil)) != sum {
		return false
	}
	return true
}
