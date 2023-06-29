package flowchart

import (
	"fmt"
	"strings"
)

type edgeInterpolation string

// Interpolation definitions for Edges as described at
// https://github.com/knsv/mermaid/issues/580#issuecomment-373929046.
// The default behaviour if no interpolation is given at all is
// InterpolationLinear.
const (
	InterpolationBasis  edgeInterpolation = `basis`
	InterpolationLinear edgeInterpolation = `linear`
)

// An EdgeStyle is used to add CSS to an Edge. It renders to a linkStyle line
// for each Edge it is associated with. Note that linkStyles will override any
// effect from the Edge's shape defintion.
// Retrieve an instance of EdgeStyle via Flowchart's EdgeStyle method, do not
// create instances directly.
type EdgeStyle struct {
	id            string            // virtual ID for lookup
	Stroke        htmlColor         // Renders to stroke:#333
	StrokeWidth   uint8             // Renders to stroke-width:2px
	StrokeDash    uint8             // Renders to stroke-dasharray:5px
	More          string            // More styles, e.g.: stroke:#333,stroke-width:1px
	Interpolation edgeInterpolation // Edge curve definition
}

// ID provides access to the EdgeStyle's readonly field id.
func (es *EdgeStyle) ID() (id string) {
	return es.id
}

// String renders this graph element to a linkStyle line.
func (es *EdgeStyle) String() (renderedElement string) {
	interpolation := ""
	if es.Interpolation != "" {
		interpolation = "interpolate " + string(es.Interpolation)
	}
	styles := []string{}
	if es.Stroke != "" {
		styles = append(styles, "stroke:"+string(es.Stroke))
	}
	if es.StrokeWidth != 1 {
		styles = append(styles, fmt.Sprintf(`stroke-width:%dpx`,
			es.StrokeWidth))
	}
	if es.StrokeDash != 0 {
		styles = append(styles, fmt.Sprintf(`stroke-dasharray:%dpx`,
			es.StrokeDash))
	}
	if es.More != "" {
		styles = append(styles, es.More)
	}
	definitions := strings.Join(styles, ",")
	if definitions == "" && interpolation == "" {
		// neutral element as a fallback to ensure empty linkStyles don't break
		// the mermaid syntax
		definitions = fmt.Sprintf(`stroke-width:%dpx`, es.StrokeWidth)
	}
	if definitions != "" && interpolation != "" {
		interpolation += " "
	}
	return `linkStyle %s ` + interpolation + definitions + "\n"
}
