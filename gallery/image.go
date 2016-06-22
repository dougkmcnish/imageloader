package gallery

import (
	"net/http"

	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//ImageUpload is the metadata for an uploaded image.
//Filename is a string representation of a generated
//UUID. The rest is self explanatory
type Image struct {
	UUID      string
	FirstName string
	LastName  string
	Email     string
	Address   string
	City      string
	State     string
	Zip       string
	Filename  string
	Width     int
	Height    int
	Published bool
}

//New creates a new ImageUpload struct. It takes a pointer to http.Request
//as an argument and returns a pointer to Data.
func NewImage(r *http.Request) *Image {
	u := &Image{}
	u.UUID = uuid.NewV4().String()
	u.FirstName = r.FormValue("fname")
	u.LastName = r.FormValue("lname")
	u.Address = r.FormValue("address")
	u.City = r.FormValue("city")
	u.State = r.FormValue("state")
	u.Zip = r.FormValue("zip")
	u.Email = r.FormValue("email")
	u.Filename = u.UUID + ".png"
	return u
}

func (u Image) Publish(session *mgo.Session) error {
	defer session.Close()
	c := session.DB("gallery").C("pictures")
	return c.Update(bson.M{"uuid": u.UUID}, bson.M{"published": true})
}

//Persist stores contents of *Upload in a MongoDB
//database. It returns error.
func (u Image) Persist(session *mgo.Session) error {
	defer session.Close()
	c := session.DB("gallery").C("pictures")
	return c.Insert(u)
}
