package index

import (
	"fmt"
	"encoding/json"
	"bytes"
	"io"
	"io/ioutil"

	"code.google.com/p/go-sqlite/go1/sqlite3"
	"github.com/rwcarlsen/cas/blobdb"
	"github.com/rwcarlsen/cas/schema"
	"github.com/rwcarlsen/cas/schema/file"
)

type Index struct {
	blobdb.Interface
	sqldb *sqlite3.Conn
}

func New(db blobdb.Interface, path string) (*Index, error) {
	sqldb, err := sqlite3.Open(path)
	if err != nil {
		return nil, err
	}

	ind := &Index{
		Interface: db,
		sqldb:            sqldb,
	}
	ind.createTables()
	return ind, nil
}

func (ind *Index) Put(r io.Reader) (ref string, n int64, err error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return "", 0, err
	}

	ref, n, err = ind.Interface.Put(bytes.NewBuffer(data))
	ind.index(ref, data)

	return ref, n, err
}

func (ind *Index) createTables() {
	tblCols := "blobref TEXT, timecreated INTEGER, objref TEXT"
	cmd := fmt.Sprintf("CREATE TABLE IF NOT EXISTS tblName (%s)", tblCols)
	err := ind.sqldb.Exec(cmd)
	if err != nil {
		panic(err)
	}

	tblCols = "blobref TEXT, timecreated INTEGER, size INTEGER, path TEXT"
	cmd = fmt.Sprintf("CREATE TABLE IF NOT EXISTS fileinfo (%s)", tblCols)
	err = ind.sqldb.Exec(cmd)
	if err != nil {
		panic(err)
	}
}

func (ind *Index) index(ref string, data []byte) {
	m := &schema.Meta{}
	if err := json.Unmarshal(data, m); err != nil {
		return
	}

	err := ind.sqldb.Exec("INSERT INTO metablobs (blobref, timecreated, objref) VALUES (?, ?, ?)", ref, m.Created, m.ObjRef)
	if err != nil {
		panic(err)
	}

	fi := &file.Info{}
	if err := schema.UnmarshalProp(data, file.Property, fi); err == nil {
		ind.indexFile(ref, fi)
	}
}

func (ind *Index) indexFile(ref string, fi *file.Info) {
	err := ind.sqldb.Exec("INSERT INTO fileinfo (blobref, timecreated, size, path) VALUES (?, ?, ?, ?)", ref, fi.Created, fi.Size, fi.Path)
	if err != nil {
		panic(err)
	}
}

func (ind *Index) RecentFiles(limit int) (blobrefs []string, err error) {
	s, err := ind.sqldb.Query("SELECT blobref FROM fileinfo ORDER BY timecreated DESC LIMIT ?", limit)
	if err != nil {
		return nil, error
	}

	for err := nil; err == nil; err = s.Next() {
		ref := ""
		if err := s.Scan(&ref); err != nil {
			return nil, err
		}
		blobrefs = append(blobrefs, ref)
	}
	return blobrefs, nil
}

