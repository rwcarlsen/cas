package localdisk

import (
	"testing"
	"os"
	"bytes"
	"io/ioutil"
	"sort"
)

func TestPutGet(t *testing.T) {
	db, err := New("/tmp/mydb")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("/tmp/mydb")

	data := []byte("hello my friends")

	ref, err := db.Put(bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}

	r, err := db.Get(ref)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	retrieved, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(retrieved, data) {
		t.Fatalf("%s != %s", retrieved, data)
	}
}

func TestEnumerate(t *testing.T) {
	db, err := New("/tmp/mydb")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("/tmp/mydb")

	prefix := []byte("hello")
	refs := []string{}
	for i := 0; i < 10; i++ {
		ref, err := db.Put(bytes.NewBuffer(append(prefix, byte(i))))
		if err != nil {
			t.Fatal(err)
		}
		refs = append(refs, ref)
	}

	sort.Strings(refs)

	listed, err := db.Enumerate(refs[4], 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(listed) != 5 {
		t.Fatalf("len(listed)=%v, expected 5", len(listed))
	}
}
