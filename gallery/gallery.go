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
	FirstName string
	LastName  string
	Email     string
	Address   string
	Filename  string
	Width     int
	Height    int
	Published bool
}

//New creates a new ImageUpload struct. It takes a pointer to http.Request
//as an argument and returns a pointer to Data.
func NewImage(r *http.Request) *Image {
	uuid := uuid.NewV4().String()
	u := &Image{}
	u.FirstName = r.FormValue("fname")
	u.LastName = r.FormValue("lname")
	u.Address = r.FormValue("address")
	u.Email = r.FormValue("email")
	u.Filename = uuid + ".png"
	return u
}

func ListMatch(session *mgo.Session) ([]Image, error) {
	defer session.Close()
	c := session.DB("gallery").C("pictures")
	var results []Image
	err := c.Find(bson.M{"published": true}).All(&results)
	return results, err
}

//Persist stores contents of *Upload in a MongoDB
//database. It returns error.
func (u Image) Persist(session *mgo.Session) error {
	defer session.Close()
	c := session.DB("gallery").C("pictures")
	return c.Insert(u)
}
