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
	public := flag.String("public", "public", "Directory for storing public assets.\n\tThis is where image files will be saved.")
	uri := flag.String("dbhost", "localhost", "Mongodb hostname")
	db := flag.String("db", "gallery", "Mongo databasee name.")
	collection := flag.String("collection", "pictures", "Mongo Collection Name.")
	port := flag.String("port", ":8080", "Port to listen on")
	flag.Parse()

	wd, _ := os.Getwd()

	log.Printf("\nRuntime:\n\n")
	log.Printf("Current working directory: %v", wd)
	log.Printf("Templates: %v", *tpl)
	log.Printf("Images: %v", *public)
	log.Printf("Port: %v", *port)
	log.Printf("URI: %v", *uri)
	log.Printf("Database: %v", *db)
	log.Printf("Collection: %v", *collection)

	return &gallery.Config{*port, uint(*minw), uint(*minh), 0, 0, *tpl, *public, *uri, *db, *collection}
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
