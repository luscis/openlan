package libol

var (
	Date    string
	Version string
	Commit  string
)

func ShowVersion() {
	Info("version: %s build at: %s commit id: %s", Version, Date, Commit)
}
