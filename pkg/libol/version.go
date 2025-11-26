package libol

var (
	Date    string
	Version string
	Commit  string
)

func ShowVersion() {
	Info("Version is %s", Version)
	Info("Built   on %s", Date)
	Info("Commit  id %s", Commit)
}
