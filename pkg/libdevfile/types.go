package libdevfile

type DevfileEventType string

const (
	// PreStart is a devfile event
	PreStart DevfileEventType = "preStart"

	// PostStart is a devfile event
	PostStart DevfileEventType = "postStart"

	// PreStop is a devfile event
	PreStop DevfileEventType = "preStop"

	// PostStop is a devfile event
	PostStop DevfileEventType = "postStop"
)
