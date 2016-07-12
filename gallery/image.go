package gallery

import (
	"net/http"

	"github.com/satori/go.uuid"
	valid "gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	valid.SetFieldsRequiredByDefault(true)
}

//ImageUpload is the metadata for an uploaded image.
//Filename is a string representation of a generated
//UUID. The rest is self explanatory
type Image struct {
	UUID      string `valid:"required,uuidv4"`
	FirstName string `valid:"required~First name is required.,alpha~First Name: Invalid characters."`
	LastName  string `valid:"required~Last name is required.,alpha"`
	Email     string `valid:"required~Email address is required,email!Invalid email address."`
	Address   string `valid:"optional,ascii~Invalid address."`
	City      string `valid:"required~City is requried.,alpha~Invalid city."`
	State     string `valid:"required~State required.,ascii,length(2|2)~Invalid state"`
	Zip       string `valid:"required~Zip code required.,matches(^[0-9]{5}$)~Invalid ZIP code."`
	Filename  string `valid:"required"`
	Width     int    `valid:"-"`
	Height    int    `valid:"-"`
	Published bool   `valid:"-"`
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
	u.Width = 0
	u.Height = 0
	u.Filename = u.UUID + ".png"

	if u.Address == "" {
		u.Address = "No address specified."
	}

	return u
}

func (u Image) Publish(session *mgo.Session) error {
	defer session.Close()
	c := session.DB("gallery").C("pictures")
	return c.Update(bson.M{"uuid": u.UUID}, bson.M{"published": true})
}



func (u Image) Valid() (bool, error) {
	return valid.ValidateStruct(u)
}
