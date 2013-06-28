package blobdb

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var DefaultSpecsPath = filepath.Join(os.Getenv("HOME"), ".blobdbs")

// Params is used to hold/specify details required for the initialization
// of standard and custom blobdb's.
type Params map[string]string

// TypeFunc is an abstraction allowing package-external blobdb's to be
// handled by this package. A TypeFunc instance should return a
// ready-to-use blobdb initialized with Params.
type TypeFunc func(Params) (Interface, error)

// Type specifies a unique kind of blobdb. There is a one-to-one
// correspondence between blobdb Types and TypeFunc's.
type Type string

var types = map[Type]TypeFunc{}

// Register enables blobdb's of type t to be created by Make functions and
// methods in this package.
func Register(t Type, fn TypeFunc) {
	types[t] = fn
}

// Make creates a blobdb of type t, initialized with the given params.
// params must contain required information for the specified blobdb type.
// An error is returned if t is an unregistered type or if params do not
// contain all pertinent information to initialize a blobdb of type t.
func Make(t Type, params Params) (Interface, error) {
	if fn, ok := types[t]; ok {
		return fn(params)
	}
	return nil, fmt.Errorf("blobdb: Invalid blobdb type %v", t)
}

// Spec is a convenient way to group a specific set of config Params for a
// blobdb together with its corresponding Type.
type Spec struct {
	DbType   Type
	DbParams Params
}

// Make creates a blobdb from the Spec. This is a shortcut for the Make
// function.
func (s *Spec) Make() (Interface, error) {
	return Make(s.DbType, s.DbParams)
}

// SpecList is a convenient way to manage multiple blobdb Spec's together
// as a group (e.g. saving to/from a config file, etc).
type SpecList struct {
	list map[string]*Spec
}

// LoadSpecList creates a SpecList by decoding JSON data from r.
func LoadSpecList(r io.Reader) (*SpecList, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	list := map[string]*Spec{}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, prettySyntaxError(string(data), err)
	}

	return &SpecList{list: list}, nil
}

// Save writes the SpecList in JSON format to w.
func (s *SpecList) Save(w io.Writer) error {
	data, err := json.Marshal(s.list)
	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}

// Get retrieves the named Spec. It returns nil if name is not found.
func (s *SpecList) Get(name string) *Spec {
	s.init()
	return s.list[name]
}

// Set adds a new Spec with the given name to the specset. If name is
// already in the specset, it is overwritten.
func (s *SpecList) Set(name string, spec *Spec) {
	s.init()
	s.list[name] = spec
}

// Make creates a blobdb from Spec identified by name. This is a shortcut
// for ".Get(...).Make(...)".
func (s *SpecList) Make(name string) (Interface, error) {
	s.init()
	if spec, ok := s.list[name]; ok {
		return spec.Make()
	}
	return nil, fmt.Errorf("blobdb: name '%v' not found in SpecList", name)
}

func (s *SpecList) init() {
	if s.list == nil {
		s.list = make(map[string]*Spec)
	}
}

func prettySyntaxError(js string, err error) error {
	syntax, ok := err.(*json.SyntaxError)
	if !ok {
		return err
	}

	start, end := strings.LastIndex(js[:syntax.Offset],
		"\n")+1, len(js)
	if idx := strings.Index(js[start:], "\n"); idx >= 0 {
		end = start + idx
	}

	line, pos := strings.Count(js[:start], "\n"), int(syntax.Offset)-start-1

	msg := fmt.Sprintf("Error in line %d: %s\n", line+1, err)
	msg += fmt.Sprintf("%s\n%s^", js[start:end], strings.Repeat("", pos))
	return pretty(msg)
}

type pretty string

func (p pretty) Error() string {
	return string(p)
}
