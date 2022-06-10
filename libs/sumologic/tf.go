package sumologic

import (
	"bytes"
	"embed"
	"github.com/OpenSLO/slogen/libs/specs"
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
