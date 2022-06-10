package libs

import (
	"fmt"
	oslo "github.com/OpenSLO/oslo/pkg/manifest/v1"
	"github.com/OpenSLO/oslo/pkg/manifest/v1alpha"
	"github.com/OpenSLO/slogen/libs/specs"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	BudgetingMethodNameTimeSlices  = "Timeslices"
	BudgetingMethodNameOccurrences = "Occurrences"
)

type SLOv1Alpha struct {
	*v1alpha.SLO   `yaml:",inline"`
	Labels         map[string]string `yaml:"labels,omitempty"`
	Fields         map[string]string `yaml:"fields,omitempty"`
	Alerts         Alerts            `yaml:"alerts,omitempty"`
	BurnRateAlerts []BurnRate        `yaml:"burnRateAlerts,omitempty"` // deprecated
}

func (s SLOv1Alpha) Name() string {
	return s.Metadata.Name
}

type Alerts struct {
	BurnRate []BurnRate `yaml:"burnRate,omitempty"`
}

func Parse(filename string) (*SLOMultiVerse, error) {
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	specVersion := giveOpenSLOVersion(fileContent)

	slo := &SLOMultiVerse{}

	switch specVersion {

	case oslo.APIVersion:
		slo.V1, err = parseV1(fileContent)
	case v1alpha.APIVersion:
		slo.Alpha, err = parseV1Alpha(fileContent)
	default:
		return nil, fmt.Errorf("unsupported OpenSLO spec version %s", specVersion)
	}

	return slo, err
}

func parseV1Alpha(yamlBody []byte) (*SLOv1Alpha, error) {

	var slo SLOv1Alpha
	err := yaml.Unmarshal(yamlBody, &slo)

	return &slo, err
}

func parseV1(yamlBody []byte) (*specs.OpenSLOSpec, error) {
	var spec specs.OpenSLOSpec

	err := yaml.Unmarshal(yamlBody, &spec)

	if err == nil {
		//pretty.Println(spec)
		//pretty.Println(sumologic.ConvertToSumoSLO(&spec))
	}

	return &spec, err
}

func giveOpenSLOVersion(yamlBody []byte) string {
	var m oslo.ObjectGeneric

	err := yaml.Unmarshal(yamlBody, &m)

	if err != nil {
		panic(err)
	}

	return m.APIVersion
}

func (s SLOv1Alpha) Target() float64 {
	target := s.Spec.Objectives[0].BudgetTarget

	if target == nil {
		return 0.99
	}

	return *target
}

func (s SLOv1Alpha) TimesliceTarget() float64 {
	target := s.Spec.Objectives[0].TimeSliceTarget

	if target == nil {
		return s.Target()
	}

	return *target
}

type SLOMultiVerse struct {
	Alpha *SLOv1Alpha
	V1    *specs.OpenSLOSpec
}
