package main

import (
	"flag"
	"github.com/rwcarlsen/cas/appserv"
	"github.com/rwcarlsen/cas/appserv/fupload"
	"github.com/rwcarlsen/cas/appserv/notedrop"
	"github.com/rwcarlsen/cas/appserv/pics"
	"github.com/rwcarlsen/cas/appserv/recent"
	"log"
)

var static = flag.String("static", "", "the app server looks for webapp static files here")

func main() {
	flag.Parse()
	log.Println("static=", *static, "::", *static == "")
	if *static == "" {
		log.Fatal("must specify path to static files")
	}

	appserv.SetStatic(*static)

	//// add new apps by listing them here in this init func
	appserv.RegisterApp("pics", pics.Handler)
	appserv.RegisterApp("notedrop", notedrop.Handler)
	appserv.RegisterApp("recent", recent.Handler)
	appserv.RegisterApp("fupload", fupload.Handler)

	err := appserv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
