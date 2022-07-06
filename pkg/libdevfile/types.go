package libdevfile

type DevfileEventType string

const (

	// PreStop is a devfile event
	PreStop DevfileEventType = "preStop"
)

type DevfileCommands struct {
	BuildCmd string
	RunCmd   string
	DebugCmd string
}
