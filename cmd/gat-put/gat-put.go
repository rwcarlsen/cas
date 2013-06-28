package main

import (
	"flag"
	"os"
	"log"
	"path/filepath"

	"github.com/rwcarlsen/cas/blobdb"
	"github.com/rwcarlsen/cas/index"
	"github.com/rwcarlsen/cas/index/file"
)

var (
	dbname = flag.String("db", "", "name of the blobdb")
	ind    = flag.String("index", "", "path to blobdb index")
	isfile   = flag.Bool("file", true, "add file metadata to index (must set index flag)")
	fpath  = flag.String("pathref", ".", "path stored in the index is relative to this (only used if -file=true)")
)

var lg = log.New(os.Stderr, "[gat-put]: ", 0)

func main() {
	flag.Parse()

	// load database
	f, err := os.Open(blobdb.DefaultSpecsPath)
	if err != nil {
		lg.Fatal(err)
	}

	specs, err := blobdb.LoadSpecList(f)
	if err != nil {
		lg.Fatal(err)
	}

	db, err := specs.Make(*dbname)
	if err != nil {
		lg.Fatal(err)
	}

	if *isfile {
		// load index
		indx, err := index.New(*ind)
		if err != nil {
			lg.Fatalf("Failed to access index %v", *ind)
		}

		// dump listed files as blobs
		store := &file.Store{db, indx}
		for _, path := range flag.Args() {
			mpath, err := filepath.Rel(*fpath, path)
			if err != nil {
				lg.Println(err)
				continue
			}

			f, err := os.Open(path)
			if err != nil {
				lg.Println(err)
				continue
			}

			if _, _, err := store.PutReader(mpath, f); err != nil {
				lg.Println(err)
			}
			f.Close()
		}
	} else {
		// dump files as blobs (w/o index metadata)
		for _, path := range flag.Args() {
			f, err := os.Open(path)
			if err != nil {
				lg.Println(err)
				continue
			}
			if _, _, err := db.Put(f); err != nil {
				lg.Println(err)
			}
			f.Close()
		}
	}
}
