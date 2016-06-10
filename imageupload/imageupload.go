package imageupload

import (
	"log"
	"net/http"

	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
)

//ImageUpload is the metadata for an uploaded image.
//Filename is a string representation of a generated
//UUID. The rest is self explanatory
type ImageUpload struct {
	FirstName string
	LastName  string
	Email     string
	Address   string
	Filename  string
	Width     int
	Height    int
}

//New creates a new ImageUpload struct. It takes a pointer to http.Request
//as an argument and returns an Upload.
func New(r *http.Request) ImageUpload {
	uuid := uuid.NewV4().String()
	u := ImageUpload{}
	u.FirstName = r.FormValue("fname")
	u.LastName = r.FormValue("lname")
	u.Address = r.FormValue("address")
	u.Email = r.FormValue("email")
	u.Filename = uuid + ".png"
	return u
}

//Persist stores contents of *Upload in a MongoDB
//database. It returns error.
func (u ImageUpload) Persist(session *mgo.Session) error {

	defer session.Close()
	c := session.DB("livegig").C("pictures")
	err := c.Insert(u)

	if err != nil {
		log.Fatal(err)
	}

	return err

}
