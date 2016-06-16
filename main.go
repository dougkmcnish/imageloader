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

	"github.com/easy-bot/httputil/response"
	"github.com/easy-bot/imageloader/gallery"
)

const (
	//MaxMemory            = 10 * 1024 * 1024
	//SaveError     string = "[%s] Could not save image: %s\n"
	MetaDataError string = "[%s] Could not update DB: %s\n"
	//ListFormError string = "Bad form data: f=%s, q=%s"
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
func HandleImage(r *http.Request) (*gallery.Image, error) {
	file, _, err := r.FormFile("file")

	if err != nil {
		log.Println(err)
		return nil, err
	}

	img, _, err := image.Decode(file)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	bounds := img.Bounds()

	if uint(bounds.Max.X) < c.MinWidth || uint(bounds.Max.Y) < c.MinHeight {
		return nil, errors.New("Image must be at least 480x480.")
	}

	meta := gallery.NewImage(r)
	meta.Width = bounds.Max.X
	meta.Height = bounds.Max.Y

	err = SaveImage(&img, meta.Filename)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return meta, nil

}

//SaveImage decodes a JPG/GIF/PNG file. It takes the file
//portion of a mime/multipart Form and a filename as arguments.
//Image data is checked against size constraints and the file
//is written to disk.
func SaveImage(i *image.Image, out string) error {
	outf, err := os.Create(filepath.Join(c.OutDir, out))
	if err != nil {
		log.Println(err)
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

//HandleUpload handles HTTP POST requests to '/upload'
func HandleUpload(w http.ResponseWriter, r *http.Request) {

	res := response.New()
	meta, err := HandleImage(r)

	if err != nil {
		log.Printf("Could not process image upload %s", err)
		res.Fatal(err.Error())
		res.Send(w, http.StatusInternalServerError)
		return
	}

	err = meta.Persist(dbPool.Copy())

	if err != nil {
		log.Printf(MetaDataError, meta.Filename, err)
		res.Fatal("Could not upload image.")
		res.Send(w, http.StatusInternalServerError)
		return
	}

	res.Append("Image uploaded successfully")
	res.Send(w, http.StatusOK)
}

func ListImages(w http.ResponseWriter, r *http.Request) {

	res := response.New()

	images, err := gallery.ListMatch(dbPool.Copy())

	if err != nil {
		res.Fatal("Image search failed.")
		res.Send(w, http.StatusInternalServerError)
		return
	}

	if len(images) > 0 {
		j, err := json.Marshal(images)

		if err != nil {
			res.Fatal("Internal Server Error")
			res.Send(w, http.StatusInternalServerError)
			return
		}
		res.Data = string(j)
	}

	res.Send(w, http.StatusOK)
	return

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
