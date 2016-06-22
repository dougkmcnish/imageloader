package gallery

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/abbot/go-http-auth"
	"github.com/easy-bot/httputil/response"
	"github.com/nfnt/resize"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
)

type Gallery struct {
	Config *Config
	DbPool *mgo.Session
}

type Page struct {
	Images []Image
	Title  string
}

//HandleImage extracts data from a HTML form.
//It extracts and parses the POSTed image and
//form fields and creates a PNG and its metadata.
//The image is saved to its configured web root.
//The function returns the image metadata to be
//persisted.
func (g Gallery) ProcessImage(r *http.Request) error {
	session := g.DbPool.Copy()
	defer session.Close()

	imgData := NewImage(r)

	if res, err := imgData.Valid(); !res {
		return err
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return err
	}

	i, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	bounds := i.Bounds()

	if uint(bounds.Max.X) < g.Config.MinWidth || uint(bounds.Max.Y) < g.Config.MinHeight {
		return errors.New("Image must be at least 480x480.")
	}

	imgData.Width = bounds.Max.X
	imgData.Height = bounds.Max.Y

	thumb := resize.Thumbnail(200, 150, i, resize.Lanczos3)

	if err = imgData.Persist(session); err != nil {
		return err
	}

	if err = g.SaveImage(&thumb, "thumb_"+ imgData.Filename); err != nil {
		return err
	}

	if err = g.SaveImage(&i, imgData.Filename); err != nil {
		return err
	}

	return nil

}

//SaveImage decodes a JPG/GIF/PNG file. It takes the file
//portion of a mime/multipart Form and a filename as arguments.
//Image data is checked against size constraints and the file
//is written to disk.
func (g Gallery) SaveImage(i *image.Image, out string) error {
	outf, err := os.Create(filepath.Join(g.Config.PubDir, out))
	defer outf.Close()
	if err != nil {
		return err
	}
	return png.Encode(outf, *i)
}

//HandleUpload extracts an image from multipart/form-data
//received via HTTP POST. It creates a new github.com/easy-bot/httputil/response.Body
//for later marshalling and return to the requesting client.
func (g Gallery) Upload(w http.ResponseWriter, r *http.Request) {
	res := response.New()
	err := g.ProcessImage(r)

	if err != nil {
		log.Printf("%v %v %v %v", r.RemoteAddr, r.Method, r.URL, err.Error())
		res.Fatal(err.Error())
	}

	SendResponse(w, res)
}

func (g Gallery) Publish(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	res := response.New()
	i := r.Form.Get("i")

	if i == "" {
		log.Printf("Missing user input")
		res.Fatal("UUID required")
		w.WriteHeader(http.StatusBadRequest)
		SendResponse(w, res)
		return
	}

	session := g.DbPool.Copy()
	defer session.Close()
	c := session.DB(g.Config.DB).C(g.Config.C)

	err := c.Update(bson.M{"uuid": i}, bson.M{"published": true})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Print("Could not look up images")
	}
	SendResponse(w, res)
}

func (g Gallery) Publisher(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	session := g.DbPool.Copy()
	defer session.Close()
	c := session.DB(g.Config.DB).C(g.Config.C)
	var images []Image
	err := c.Find(nil).All(&images)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Print(w, "DB lookup failed.")
		return
	}

	t := template.Must(template.ParseFiles(filepath.Join(g.Config.AssetDir, "index.html.tmpl")))
	p := &Page{Title: "Image publisher.", Images: images}

	t.Execute(w, p)
}

func (g Gallery) ListAll(w http.ResponseWriter, r *http.Request) {
	g.ImageList(w, r, nil)
}

func (g Gallery) ListPublished(w http.ResponseWriter, r *http.Request) {
	g.ImageList(w, r, bson.M{"published": true})
}

//ListImages queries MongoDB for a list of published images.
func (g Gallery) ImageList(w http.ResponseWriter, r *http.Request, q bson.M) {
	res := response.New()

	session := g.DbPool.Copy()
	defer session.Close()
	c := session.DB(g.Config.DB).C(g.Config.C)

	var images []Image
	err := c.Find(q).All(&images)

	if err != nil {
		log.Printf("%v %v %v %v", r.RemoteAddr, r.Method, r.URL, err)
		res.Fatal("Image search failed.")
	}

	if len(images) > 0 {
		files := make([]string, len(images))

		for i, e := range images {
			files[i] = e.Filename
		}

		j, err := json.Marshal(files)

		if err != nil {
			log.Printf("%v %v %v %v", r.RemoteAddr, r.Method, r.URL, err)
			res.Fatal("Could not parse image list.")
		}

		if j != nil {
			res.Data = string(j)
		}
	}

	SendResponse(w, res)
}

//SendResponse serializes a httputil.response.Body into JSON
//and sends it to the requesting process.
// Call http.ResponseWriter.WriteHeader if you need to send
// a return code other than 200.
func SendResponse(w http.ResponseWriter, res response.Body) {
	json, err := res.Json()
	if err != nil {
		log.Fatal("Could not marhal response body.")
	}
	w.Header().Add("Access-Control-Allow-Origin", "http://www.rtctel.com")
	fmt.Fprint(w, json)
}

func NewGallery(c *Config) *Gallery {
	session, err := mgo.Dial(c.DatabaseURI)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	return &Gallery{c, session}
}
