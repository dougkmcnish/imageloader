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
	"os"

	"encoding/json"

	"fmt"
	"github.com/easy-bot/httputil/response"
	"github.com/easy-bot/imageloader/gallery"
)

//Config stores runtime configuration.
//BUG(dag) Config should be serialized for later runs.
type Config struct {
	Listen      string
	MinWidth    uint
	MinHeight   uint
	MaxWidth    uint
	MaxHeight   uint
	TemplateDir string
	OutDir      string
}

var c *Config
var dbPool *mgo.Session

//HandleImage extracts data from a HTML form.
//It extracts and parses the POSTed image and
//form fields and creates a PNG and its metadata.
//The image is saved to its configured web root.
//The function returns the image metadata to be
//persisted.
func HandleImage(r *http.Request, db *mgo.Session) error {

	file, _, err := r.FormFile("file")
	if err != nil {
		return err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	bounds := img.Bounds()

	if uint(bounds.Max.X) < c.MinWidth || uint(bounds.Max.Y) < c.MinHeight {
		return errors.New("Image must be at least 480x480.")
	}

	meta := gallery.NewImage(r)
	meta.Width = bounds.Max.X
	meta.Height = bounds.Max.Y

	if err = SaveImage(&img, meta.Filename); err != nil {
		return err
	}

	if err = meta.Persist(db); err != nil {
		return err
	}

	return nil

}

//SaveImage decodes a JPG/GIF/PNG file. It takes the file
//portion of a mime/multipart Form and a filename as arguments.
//Image data is checked against size constraints and the file
//is written to disk.
func SaveImage(i *image.Image, out string) error {
	outf, err := os.Create(filepath.Join(c.OutDir, out))
	if err != nil {
		return err
	}
	return png.Encode(outf, *i)
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

func SendResponse(w http.ResponseWriter, res response.Body) {

	json, err := res.Json()
	if err != nil {
		log.Fatal("Could not marhal response body.")
	}

	fmt.Fprint(w, string(json))
}

//HandleUpload handles HTTP POST requests to '/upload'
func HandleUpload(w http.ResponseWriter, r *http.Request) {

	res := response.New()
	err := HandleImage(r, dbPool.Copy())

	if err != nil {
		log.Printf("Could not process image upload %s", err)
		res.Fatal(err.Error())
	}

	if res.Fatal {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	SendResponse(w, res)
}

func ListImages(w http.ResponseWriter, r *http.Request) {

	res := response.New()

	images, err := gallery.ListMatch(dbPool.Copy())

	if err != nil {
		log.Println(err)
		res.Fatal("Image search failed.")
	}

	if len(images) > 0 {
		j, err := json.Marshal(images)

		if err != nil {
			log.Println(err)
			res.Fatal("Could not parse image list.")
		}

		if j != nil {
			res.Data = string(j)
		}
	}

	if res.Fatal {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	SendResponse(w, res)
}

func ProcessArgs() (*Config, *mgo.Session) {
	tpl := flag.String("templates", "static", "Directory containing an index.html template.")
	minh := flag.Int("minheight", 480, "Minimum image height.")
	minw := flag.Int("minwidth", 480, "Minimum image Width")
	out := flag.String("output", "images", "Directory where images can be saved.")
	db := flag.String("db", "localhost", "Mongodb hostname")
	port := flag.String("port", ":8080", "Port to listen on")
	flag.Parse()

	dbPool = InitDb(*db)
	dbPool.SetMode(mgo.Monotonic, true)

	return &Config{*port, uint(*minw), uint(*minh), 0, 0, *tpl, *out}, dbPool
}

func main() {

	c, dbPool = ProcessArgs()

	wd, _ := os.Getwd()
	defer dbPool.Close()

	log.Printf("\nRuntime:\n\n")
	log.Printf("Current working directory: %s", wd)
	log.Printf("Templates: %s", c.TemplateDir)
	log.Printf("Images: %s", c.OutDir)
	log.Printf("Port: %s", c.Listen)

	http.HandleFunc("/images/upload", HandleUpload)
	http.HandleFunc("/images/", ListImages)
	http.Handle("/", http.FileServer(http.Dir(c.TemplateDir)))
	log.Fatal(http.ListenAndServe(c.Listen, nil))
}
