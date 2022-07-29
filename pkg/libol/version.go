package libol

var (
	Date    string
	Version string
	Commit  string
)

func init() {
	Debug("version is %s", Version)
	Debug("built on %s", Date)
}
