package specs

import v1 "github.com/OpenSLO/oslo/pkg/manifest/v1"

type OpenSLOSpec struct {
	*v1.SLO `yaml:",inline"`
}
