package libs

import (
	"github.com/OpenSLO/oslo/pkg/manifest/v1alpha"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	BudgetingMethodNameTimeSlices  = "Timeslices"
	BudgetingMethodNameOccurrences = "Occurrences"
)

type SLO struct {
	*v1alpha.SLO   `yaml:",inline"`
	Labels         map[string]string `yaml:"labels,omitempty"`
	Fields         map[string]string `yaml:"fields,omitempty"`
	Alerts         Alerts            `yaml:"alerts,omitempty"`
	BurnRateAlerts []BurnRate        `yaml:"burnRateAlerts,omitempty"` // deprecated
}

func (s SLO) Name() string {
	return s.Metadata.Name
}

type Alerts struct {
	BurnRate []BurnRate `yaml:"burnRate,omitempty"`
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
	target := s.Spec.Objectives[0].BudgetTarget

	if target == nil {
		return 0.99
	}

	return *target
}

func (s SLO) TimesliceTarget() float64 {
	target := s.Spec.Objectives[0].TimeSliceTarget

	if target == nil {
		return s.Target()
	}

	return *target
}
