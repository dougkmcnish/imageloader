package gallery

import (
	"net/http"

	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/validator.v2"
)

//ImageUpload is the metadata for an uploaded image.
//Filename is a string representation of a generated
//UUID. The rest is self explanatory
type Image struct {
	UUID      string `validate:"min=36,max=36,rexexp=^[0-9][a-f]-+$"`
	FirstName string `validate:"nonzero"`
	LastName  string `validate:"nonzero"`
	Email     string `validate:"nonzero,rexexp=^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$"`
	Address   string `validate:"nonzero"`
	City      string `validate:"nonzero"`
	State     string `validate:"nonzero"`
	Zip       string `validate:"nonzero,rexexp=[0-9]{5}"`
	Filename  string `validate:"nonzero"`
	Width     int    `validate:"nonzero"`
	Height    int    `validate:"nonzero"`
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
	if err := u.Validate(); err != nil {
		return err
	}
	return c.Insert(u)
}

func (u Image) Validate() error {
	if u.Address == "" {
		u.Address = "No address specified."
	}
	return validator.Validate(u)
}
