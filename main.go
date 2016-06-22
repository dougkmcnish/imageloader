// Uploader provides a web service for uploading
// images to a web server and storing metadata
// in a MongoDB database.
package main

import (
	"log"
	"net/http"

	"flag"
	"os"
	auth "github.com/abbot/go-http-auth"
	"github.com/easy-bot/imageloader/gallery"
)

func Log(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v %v %v", r.RemoteAddr, r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}

func ProcessArgs() *gallery.Config {
	tpl := flag.String("templates", "static", "Directory containing an index.html.tmpl template.")
	minh := flag.Int("minheight", 480, "Minimum image height.")
	minw := flag.Int("minwidth", 480, "Minimum image Width")
	public := flag.String("output", "images", "Directory where images can be saved.")
	db := flag.String("db", "localhost", "Mongodb hostname")
	port := flag.String("port", ":8080", "Port to listen on")
	flag.Parse()

	wd, _ := os.Getwd()

	log.Printf("\nRuntime:\n\n")
	log.Printf("Current working directory: %s", wd)
	log.Printf("Templates: %s", *tpl)
	log.Printf("Images: %s", *public)
	log.Printf("Port: %s", *port)
	log.Printf("Database: %s", *db)

	return &gallery.Config{*port, uint(*minw), uint(*minh), 0, 0, *tpl, *public, *db}
}

func main() {
	c := ProcessArgs()
	g := gallery.NewGallery(c)
	defer g.DbPool.Close()

	// read .htpasswd
	htpasswd := auth.HtpasswdFileProvider("./.htpasswd")
	authenticator := auth.NewBasicAuthenticator("Gallery", htpasswd)

	//Routes
	http.HandleFunc("/gallery/upload/", g.Upload)
	http.HandleFunc("/gallery/published/", g.ListPublished)
	http.HandleFunc("/gallery/", g.ListAll)
	http.HandleFunc("/gallery/view/", authenticator.Wrap(g.Publisher))
	http.HandleFunc("/gallery/publish/", authenticator.Wrap(g.Publish))

	http.Handle("/", http.FileServer(http.Dir(c.TemplateDir)))
	log.Fatal(http.ListenAndServe(c.Listen, Log(http.DefaultServeMux)))
}
