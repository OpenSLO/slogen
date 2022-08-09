package specs

import "github.com/agaurav/oslo/pkg/manifest/v1"

type OpenSLOSpec struct {
	*v1.SLO `yaml:",inline"`
}
