package index

import (
	"fmt"
	"time"

	"code.google.com/p/go-sqlite/go1/sqlite3"
)

// A key-value index for a blob database.
type Index struct {
	sqldb *sqlite3.Conn
}

func New(path string) (*Index, error) {
	sqldb, err := sqlite3.Open(path)
	if err != nil {
		return nil, err
	}

	ind := &Index{
		sqldb: sqldb,
	}
	ind.createTables()
	return ind, nil
}

// Set creates a key-value meta-data entry in the index for the blob specified
// by ref.  Entry changes are timestamped and never overwritten.
func (ind *Index) Set(ref, key, val string) error {
	sql := "INSERT INTO blobindex (blobref, timestamp, key, value) VALUES (?, ?, ?, ?)"
	return ind.sqldb.Exec(sql, ref, time.Now(), key, val)
}

func (ind *Index) createTables() {
	tblCols := "blobref TEXT, timestamp INTEGER, key TEXT, value TEXT"
	cmd := fmt.Sprintf("CREATE TABLE IF NOT EXISTS blobindex (%s)", tblCols)
	err := ind.sqldb.Exec(cmd)
	if err != nil {
		panic(err)
	}
}

func (ind *Index) query(sql string) (blobrefs []string, err error) {
	s, err := ind.sqldb.Query(sql)
	if err != nil {
		return nil, err
	}

	for err = nil; err == nil; err = s.Next() {
		ref := ""
		if err := s.Scan(&ref); err != nil {
			return nil, err
		}
		blobrefs = append(blobrefs, ref)
	}
	return blobrefs, nil
}

// FindExact returns a list of all blobrefs that have the key-value pair with
// values exactly as specified.  Returns at most the newest limit blobrefs.
func (ind *Index) FindExact(key, val string, limit int) (blobrefs []string, err error) {
	sql := fmt.Sprintf("SELECT DISTINCT blobref FROM blobindex WHERE key=%s AND value=%s LIMIT %v ORDER BY timestamp DESC",
		key, val, limit)
	return ind.query(sql)
}

// Find returns a list of all blobrefs that have the specified key with a value
// matching valpattern.  Returns at most the newest limit blobrefs
func (ind *Index) Find(key, valpattern string, limit int) (blobrefs []string, err error) {
	sql := fmt.Sprintf("SELECT DISTINCT blobref FROM blobindex WHERE key=%s AND value LIKE %s LIMIT %v ORDER BY timestamp DESC",
		key, valpattern, limit)
	return ind.query(sql)
}

// FindExactCurr returns a list of all blobrefs that have the key-value pair
// with values exactly as specified for the most recent value of key.  Returns
// at most the newest limit blobrefs.
func (ind *Index) FindExactCurr(key, val string, limit int) (blobrefs []string, err error) {
	s := `
		SELECT DISTINCT blobref FROM blobindex a WHERE
			key=%s AND value=%s AND timestamp = (
				SELECT MAX(timestamp) FROM blobindex b
				WHERE b.key=%s AND b.value=%s AND a.blobref=b.blobref
			)
		LIMIT %v ORDER BY timestamp DESC`
	sql := fmt.Sprintf(s, key, val, key, val, limit)
	return ind.query(sql)
}

// FindCurr returns a list of all blobrefs that have the specified key with a
// value matching valpattern for the most recent value of key.  Returns at most
// the newest limit blobrefs
func (ind *Index) FindCurr(key, valpattern string, limit int) (blobrefs []string, err error) {
	s := `
		SELECT DISTINCT blobref FROM blobindex a WHERE
			key=%s AND value LIKE %s AND timestamp = (
				SELECT MAX(timestamp) FROM blobindex b
				WHERE b.key=%s AND b.value LIKE %s AND a.blobref=b.blobref
			)
		LIMIT %v ORDER BY timestamp DESC`
	sql := fmt.Sprintf(s, key, valpattern, key, valpattern, limit)
	return ind.query(sql)
}

type Entry struct {
	Timestamp time.Time
	Key       string
	Value     string
}

// Info returns all meta-data entries from the index for the specified blobref
// up to the limit newest entries.
func (ind *Index) Info(blobref string, limit int) ([]*Entry, error) {
	sql := "SELECT timestamp, key, value FROM blobindex WHERE blobref=? ORDERED BY timestamp DESC LIMIT ?"
	s, err := ind.sqldb.Query(sql, blobref, limit)
	if err != nil {
		return nil, err
	}

	ents := make([]*Entry, 0)
	for err = nil; err == nil; err = s.Next() {
		ent := &Entry{}
		if err := s.Scan(&ent.Timestamp, &ent.Key, &ent.Value); err != nil {
			return nil, err
		}
		ents = append(ents, ent)
	}
	return ents, nil
}
