package schema

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"bytes"
)

type Kind string

const (
	Tobject Kind = "object"
	Tmeta        = "meta"
)

type Object struct {
	Type    Kind
	Created time.Time
	Notes   string
	Rand    string
}

func NewObject(notes string) io.Reader {
	rb := make([]byte, 32)
	rand.Read(rb)
	obj := &Object{
		Type:    Tobject,
		Created: time.Now(),
		Notes:   notes,
		Rand:    fmt.Sprintf("%x", rb),
	}

	data, err := Marshal(obj)
	if err != nil {
		panic(err)
	}
	return bytes.NewBuffer(data)
}

type Meta struct {
	Type    Kind
	Created time.Time
	ObjRef  string
	Props   map[string]interface{}
}

func NewMeta(objRef string) *Meta {
	return &Meta{
		Type:       Tmeta,
		Created:    time.Now(),
		ObjectRef:  objRef,
		Props: map[string]interface{}{},
	}
}

func Marshal(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "\t")
}

func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func UnmarshalProp(data []byte, prop string, v interface{}) error {
	meta := &Meta{}
	if err := json.Unmarshal(data, meta); err != nil {
		return err
	}

	data, err = json.Marshal(meta.Props[prop])
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func addField(field, val string, data []byte) []byte {
	s := string(data)
	trimmed := strings.TrimSpace(s)
	if len(trimmed) == 0 || trimmed[len(trimmed)-1] != '}' {
		panic("blob: json blob marshaling is broken")
	}
	trimmed = trimmed[:len(trimmed)-1]
	trimmed = strings.TrimSpace(trimmed)
	trimmed += ",\n\t\"" + field + "\":\"" + val + "\"\n}\n"
	return []byte(trimmed)
}
