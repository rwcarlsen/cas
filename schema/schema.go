package schema

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Operation string

const (
	PropSet Operation = "SetProperty"
	PropDel           = "DelProperty"
)

type Object struct {
	Created time.Time
	Meta    string
	Random  string
}

func NewObject(meta string) *Object {
	rb := make([]byte, 32)
	rand.Read(rb)
	return &Object{time.Now(), meta, fmt.Sprintf("%x", rb)}
}

type Mutation struct {
	Created   time.Time
	ObjectRef string
	Property  string
	Op        Operation
	Value     string
}

type mutSort []*Mutation

func (m mutSort) Len() int {
	return len(m)
}
func (m mutSort) Less(i, j int) bool {
	return m[j].Created.After(m[i].Created)
}
func (m mutSort) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func SetProp(objRef string, prop, val string) *Mutation {
	return &Mutation{time.Now(), objRef, prop, PropSet, val}
}

func DelProp(objRef string, prop string) *Mutation {
	return &Mutation{time.Now(), objRef, prop, PropDel, ""}
}

func CurrProperties(muts []*Mutation) map[string]string {
	props := map[string]string{}
	sort.Sort(mutSort(muts))
	for _, m := range muts {
		if m.Op == PropSet {
			props[m.Property] = m.Value
		} else if m.Op == PropDel {
			delete(props, m.Property)
		}
	}
	return props
}

func Marshal(v interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Unmarshal(data []byte, v interface{}) error {
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
