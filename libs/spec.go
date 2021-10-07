package libs

import (
	"github.com/OpenSLO/oslo/pkg/manifest/v1alpha"
	"gopkg.in/yaml.v3"
	"os"
)

type SLO struct {
	*v1alpha.SLO   `yaml:",inline"`
	BurnRateAlerts []BurnRate        `yaml:"burnRateAlerts,omitempty"`
	Labels         map[string]string `yaml:"labels,omitempty"`
	Fields         map[string]string `yaml:"fields,omitempty"`
	ViewName       string            `yaml:"viewName"`
}

func Parse(filename string) (*SLO, error) {
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var slo SLO
	err = yaml.Unmarshal(fileContent, &slo)

	return &slo, err
}

func (s SLO) Target() float64 {
	return *(s.Spec.Objectives[0].BudgetTarget)
}
