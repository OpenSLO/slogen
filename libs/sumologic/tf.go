package sumologic

import (
	"bytes"
	"embed"
	"github.com/OpenSLO/slogen/libs/specs"
	oslo "github.com/agaurav/oslo/pkg/manifest/v1"
	"text/template"
)

//go:embed templates/*.tf.gotf
var tmplFiles embed.FS

var tfTemplates *template.Template

const (
	SLOTmplName        = "slo.tf.gotf"
	SLOFolderTmplName  = "slo-folders.tf.gotf"
	SLOMonitorTmplName = "slo-monitors.tf.gotf"
)

func init() {
	var err error

	tfTemplates, err = template.ParseFS(tmplFiles, "templates/*.tf.gotf")
	if err != nil {
		panic(err)
	}
}

func GiveTerraform(apMap map[string]oslo.AlertPolicy, ntMap map[string]oslo.AlertNotificationTarget,
	slo specs.OpenSLOSpec) (string, string, error) {
	sloStr, err := GiveSLOTerraform(slo)

	if err != nil {
		return "", "", err
	}

	monitorsStr, err := GiveSLOMonitorTerraform(apMap, ntMap, slo)

	return sloStr, monitorsStr, err
}

func GiveSLOTerraform(s specs.OpenSLOSpec) (string, error) {

	sumoSLO, err := ConvertToSumoSLO(s)

	if err != nil {
		return "", err
	}

	tmpl := tfTemplates.Lookup(SLOTmplName)

	buff := &bytes.Buffer{}
	err = tmpl.Execute(buff, sumoSLO)

	if err != nil {
		return "", err
	}

	return buff.String(), nil
}

func GiveMonitorTerraform(mons []SLOMonitor) (string, error) {

	tmpl := tfTemplates.Lookup(SLOMonitorTmplName)

	buff := &bytes.Buffer{}
	err := tmpl.Execute(buff, mons)

	if err != nil {
		return "", err
	}

	return buff.String(), nil
}

func IsSource(slo specs.OpenSLOSpec) bool {
	indicator := slo.Spec.Indicator

	sourceType := ""

	if indicator.Spec.RatioMetric != nil {
		sourceType = indicator.Spec.RatioMetric.Total.MetricSource.Type
	}

	if indicator.Spec.ThresholdMetric != nil {
		sourceType = indicator.Spec.ThresholdMetric.MetricSource.Type
	}

	return sourceType == SourceTypeLogs || sourceType == SourceTypeMetrics
}
