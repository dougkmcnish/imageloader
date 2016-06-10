// Uploader provides a web service for uploading
// images to a web server and storing metadata
// in a MongoDB database.
package main

import (
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"net/http"
	"path/filepath"

	"gopkg.in/mgo.v2"

	"flag"
	"mime/multipart"
	"os"

	"github.com/easy-bot/httputil/response"
	"github.com/easy-bot/imageloader/imageupload"
)

// MaxMemory is the maximum size of our form data.
// Set to 10MB
const MaxMemory = 10 * 1024 * 1024

//Config stores runtime configuration.
//BUG(dag) Config should be serialized for later runs.
type Config struct {
	MinWidth    uint
	MinHeight   uint
	MaxWidth    uint
	MaxHeight   uint
	TemplateDir string
	OutDir      string
}

var c *Config
var dbPool *mgo.Session

//SaveImage decodes a JPG/GIF/PNG file. It takes the file
//portion of a mime/multipart Form and a filename as arguments.
//Image data is checked against size constraints and the file
//is written to disk.
func SaveImage(f multipart.File, out string) error {
	img, _, err := image.Decode(f)

	if err != nil {
		log.Println(err)
		return err
	}

	bounds := img.Bounds()
	if uint(bounds.Max.X) < c.MinWidth || uint(bounds.Max.Y) < c.MinHeight {
		return errors.New("Image must be at least 480x480.")
	}

	outf, err := os.Create(filepath.Join(c.OutDir, out))
	return png.Encode(outf, img)
}

//InitDb attempts a connection to a MongoDB database.
//It returns a *mgo.Session if successful and panics on failure.
func InitDb(h string) *mgo.Session {
	session, err := mgo.Dial(h)

	if err != nil {
		panic(err)
	}

	return session
}

//HandleUpload handles HTTP POST requests to '/upload'
func HandleUpload(w http.ResponseWriter, r *http.Request) {

	file, _, err := r.FormFile("file")

	u := imageupload.New(r)
	res := response.NewResponseBody()
	err = SaveImage(file, u.Filename)

	if err != nil {
		log.Printf("[%s] Could not save image: %s\n", u.Filename, err)
		res.Fatal(err.Error())
		res.Send(w, http.StatusInternalServerError)
		return
	}

	err = u.Persist(dbPool.Copy())

	if err != nil {
		log.Printf("[%s] Could not update DB: %s\n", u.Filename, err)
		res.Fatal("Could not upload image.")
		res.Send(w, http.StatusInternalServerError)
		return
	}

	res.Append("Image uploaded successfully")

	res.Send(w, http.StatusOK)
}

func main() {

	tpl := flag.String("templates", "static", "Directory containing an index.html template.")
	out := flag.String("output", "images", "Directory where images can be saved.")
	db := flag.String("db", "images", "Directory where images can be saved.")
	port := flag.String("port", ":8080", "Port to listen on")

	flag.Parse()

	t, _ := filepath.Abs(*tpl)
	o, _ := filepath.Abs(*out)
	c = &Config{480, 480, 0, 0, t, o}

	wd, _ := os.Getwd()

	dbPool = InitDb(*db)
	defer dbPool.Close()

	log.Printf("\nRuntime:\n\n")
	log.Printf("Current working directory: %s", wd)
	log.Printf("Templates: %s", c.TemplateDir)
	log.Printf("Images: %s", c.OutDir)
	log.Printf("Port: %s", *port)

	http.HandleFunc("/upload", HandleUpload)
	http.Handle("/", http.FileServer(http.Dir(c.TemplateDir)))
	log.Fatal(http.ListenAndServe(*port, nil))
}
