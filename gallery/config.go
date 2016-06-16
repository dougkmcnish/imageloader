package gallery

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
	Database    string
}
