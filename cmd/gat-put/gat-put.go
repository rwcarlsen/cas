package main

import (
	"flag"
	"log"
	"os"

	"github.com/rwcarlsen/cas/index/file"
)

var lg = log.New(os.Stderr, "[gat-put]: ", 0)

var (
	ind  = flag.String("index", "", "path to blobdb index")
	file = flag.String("file", true, "add file metadata to index (must set index flag)")
)

func main() {
	flag.Parse()

	// load database
	//db :=

	if *file {
		indx, err := index.New(*ind)
		if err != nil {
			lg.Fatalf("Failed to access index %v", *ind)
		}
		store := &file.Store{db, indx}
		for _, path := range flag.Args() {
			if _, err := store.PutPath(path); err != nil {
				lg.Println(err)
			}
		}
	} else {
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
