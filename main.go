// Uploader provides a web service for uploading
// images to a web server and storing metadata
// in a MongoDB database.
package main

import (
	"log"
	"net/http"

	"flag"
	"os"

	"github.com/easy-bot/imageloader/gallery"
)

func ProcessArgs() *gallery.Config {
	tpl := flag.String("templates", "static", "Directory containing an index.html template.")
	minh := flag.Int("minheight", 480, "Minimum image height.")
	minw := flag.Int("minwidth", 480, "Minimum image Width")
	out := flag.String("output", "images", "Directory where images can be saved.")
	db := flag.String("db", "localhost", "Mongodb hostname")
	port := flag.String("port", ":8080", "Port to listen on")
	flag.Parse()

	wd, _ := os.Getwd()

	log.Printf("\nRuntime:\n\n")
	log.Printf("Current working directory: %s", wd)
	log.Printf("Templates: %s", *tpl)
	log.Printf("Images: %s", *out)
	log.Printf("Port: %s", *port)
	log.Printf("Database: %s", *db)

	return &gallery.Config{*port, uint(*minw), uint(*minh), 0, 0, *tpl, *out, *db}
}

func main() {
	c := ProcessArgs()
	g := gallery.NewGallery(c)
	defer g.DbPool.Close()

	http.HandleFunc("/images/upload", g.HandleUpload)
	http.HandleFunc("/images/", g.ListImages)
	http.Handle("/", http.FileServer(http.Dir(c.TemplateDir)))
	log.Fatal(http.ListenAndServe(c.Listen, nil))
}
