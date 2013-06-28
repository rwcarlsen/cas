package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/rwcarlsen/cas/index"
	"github.com/rwcarlsen/cas/index/file"
)

var (
	ind     = flag.String("index", "", "path to blobdb index")
	max     = flag.Int("max", 0, "maximum number of results to retrieve")
	prefix  = flag.String("path", "", "find with a specific path")
	showOld = flag.Bool("hist", false, "include previous versions in results")
)

var lg = log.New(os.Stderr, "[gat-find]: ", 0)

func main() {
	flag.Parse()

	indx, err := index.New(*ind)
	if err != nil {
		lg.Fatalf("Failed to access index %v", *ind)
	}

	*prefix += "%"
	if !path.IsAbs(*prefix) {
		*prefix = "%" + *prefix
	}

	var refs []string
	if *showOld {
		refs, err = indx.Find(file.Path, *prefix, *max)
	} else {
		refs, err = indx.Find(file.Path, *prefix, *max)
	}

	if err != nil {
		lg.Fatal(err)
	}

	for _, ref := range refs {
		fmt.Println(ref)
	}
}
