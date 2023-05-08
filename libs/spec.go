package libs

import (
	"fmt"
	"os"

	oslo "github.com/OpenSLO/oslo/pkg/manifest/v1"
	"github.com/OpenSLO/oslo/pkg/manifest/v1alpha"
	"github.com/OpenSLO/slogen/libs/specs"
	"gopkg.in/yaml.v3"
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

	sloHeaders := giveOpenSLOVersion(fileContent)

	slo := &SLOMultiVerse{
		ConfigPath: filename,
	}

	switch sloHeaders.APIVersion {

	case oslo.APIVersion:
		err = parseV1(fileContent, sloHeaders, slo)
	case v1alpha.APIVersion:
		slo.SLOAlpha, err = parseV1Alpha(fileContent)
	default:
		return nil, fmt.Errorf("unsupported OpenSLO spec version %s", sloHeaders.APIVersion)
	}

	return slo, err
}

func parseV1Alpha(yamlBody []byte) (*SLOv1Alpha, error) {

	var slo SLOv1Alpha
	err := yaml.Unmarshal(yamlBody, &slo)

	return &slo, err
}

func parseV1(yamlBody []byte, headers oslo.ObjectGeneric, sloM *SLOMultiVerse) error {

	var err error
	switch headers.Kind {
	case oslo.KindSLO:
		var spec specs.OpenSLOSpec
		err = yaml.Unmarshal(yamlBody, &spec)
		sloM.SLO = &spec
	case oslo.KindAlertPolicy:
		var spec oslo.AlertPolicy
		err = yaml.Unmarshal(yamlBody, &spec)
		sloM.AlertPolicy = &spec
	case oslo.KindAlertCondition:
		var spec oslo.AlertCondition
		err = yaml.Unmarshal(yamlBody, &spec)
		sloM.AlertCondition = &spec
	case oslo.KindAlertNotificationTarget:
		var spec oslo.AlertNotificationTarget
		err = yaml.Unmarshal(yamlBody, &spec)
		sloM.AlertNotificationTarget = &spec
	}

	return err
}

func giveOpenSLOVersion(yamlBody []byte) oslo.ObjectGeneric {
	var m oslo.ObjectGeneric

	err := yaml.Unmarshal(yamlBody, &m)

	if err != nil {
		panic(err)
	}

	return m
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
	ConfigPath              string
	SLOAlpha                *SLOv1Alpha
	SLO                     *specs.OpenSLOSpec
	AlertPolicy             *oslo.AlertPolicy
	AlertCondition          *oslo.AlertCondition
	AlertNotificationTarget *oslo.AlertNotificationTarget
}
