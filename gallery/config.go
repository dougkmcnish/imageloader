package gallery

//Config stores runtime configuration.
//BUG(dag) Config should be serialized for later runs.
type Config struct {
	Port        string
	MinWidth    uint
	MinHeight   uint
	MaxWidth    uint
	MaxHeight   uint
	AssetDir    string
	PubDir      string
	DatabaseURI string
	DB          string
	C           string
}
