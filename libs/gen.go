package libs

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/OpenSLO/slogen/libs/specs"
	"github.com/OpenSLO/slogen/libs/sumologic"
	"github.com/kr/pretty"
	"os"
	"path/filepath"
	"sort"
	"text/template"
)

var tfTemplates *template.Template

const NameDashboardTmpl = "dashboard.tf.gotf"
const NameViewTmpl = "sched-view.tf.gotf"
const NameMonitorTmpl = "monitor.tf.gotf"
const NameMainTmpl = "main.tf.gotf"
const NameModuleTmpl = "module_interface.tf.gotf"
const NameDashFolderTmpl = "dash-folders.tf.gotf"
const NameMonitorFolderTmpl = "monitor-folders.tf.gotf"
const NameGlobalTrackerTmpl = "global-tracker.tf.gotf"
const NameServiceTrackerTmpl = "service-overview.tf.gotf"

//go:embed templates/terraform/**/*.tf.gotf
var tmplFiles embed.FS

const (
	BuildFolder      = "build"
	ViewsFolder      = "views"
	MonitorsFolder   = "monitors"
	DashboardsFolder = "dashboards"
)

type GenConf struct {
	OutDir        string
	DashFolder    string
	MonitorFolder string
	ViewPrefix    string
	DoPlan        bool
	DoApply       bool
	IgnoreError   bool
	Clean         bool
	AsModule      bool
	UseViewHash   bool
}

func init() {
	var err error

	tfTemplates, err = template.ParseFS(tmplFiles, "templates/terraform/sumologic/**")
	if err != nil {
		panic(err)
	}
}

const ViewPrefix = "slogen_tf"

// GenTerraform for all openslo files in the path
func GenTerraform(slosMv map[string]*SLOMultiVerse, c GenConf) (string, error) {

	slosAlpha, slosV1 := splitMultiVerse(slosMv)
	err := SetupOutDir(c)
	if err != nil {
		BadResult("error setting up path : %s", err)
		return "", err
	}

	//pretty.Println(slosV1)

	genTerraformForV1(slosV1, c)
	return genTerraformForAlpha(slosAlpha, c)
}

func genTerraformForAlpha(slosAlpha map[string]*SLOv1Alpha, c GenConf) (string, error) {
	var err error

	useViewID = c.UseViewHash
	srvMap := map[string]bool{}

	for path, s := range slosAlpha {

		srvMap[s.Spec.Service] = true

		tmplToDo := []string{NameViewTmpl, NameDashboardTmpl, NameMonitorTmpl}

		GoodInfo("\nGenerating tf for SLO : %s\n", s.Metadata.Name)

		for _, t := range tmplToDo {
			err = ExecSLOTmpl(t, *s, c.OutDir)
			if err != nil {
				return path, err
			}
			log.Infof("completed stage %s", t)
		}
	}

	srvList := GiveKeys(srvMap)
	sort.Strings(srvList)

	err = GenFoldersTF(srvList, c.OutDir)
	if err != nil {
		return "", err
	}

	err = GenOverviewTF(slosAlpha, c)
	return "", err
}

func genTerraformForV1(slos map[string]*specs.OpenSLOSpec, c GenConf) error {

	v1Path := filepath.Join(c.OutDir, "sumologic")
	err := EnsureDir(v1Path, c.Clean)

	if err != nil {
		return err
	}

	for _, slo := range slos {
		sumoSLO, err := sumologic.GiveSLOTerraform(*slo)

		if err != nil {
			BadUResult(err.Error())
			return err
		}

		pretty.Println(sumoSLO)
		err = os.WriteFile(filepath.Join(v1Path, fmt.Sprintf("slo_%s.tf", slo.Metadata.Name)), []byte(sumoSLO), 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func splitMultiVerse(slos map[string]*SLOMultiVerse) (map[string]*SLOv1Alpha, map[string]*specs.OpenSLOSpec) {

	alpha := map[string]*SLOv1Alpha{}
	v1 := map[string]*specs.OpenSLOSpec{}

	for path, sloMv := range slos {
		if sloMv.Alpha != nil {
			alpha[path] = sloMv.Alpha
		}
		if sloMv.V1 != nil {
			v1[path] = sloMv.V1
		}
	}

	return alpha, v1
}

func ExecSLOTmpl(tmplName string, slo SLOv1Alpha, outDir string) error {

	var tfFilePath string
	switch tmplName {
	case NameViewTmpl:
		vc, err := ViewConfigFromSLO(slo)
		if err != nil {
			return err
		}
		tfFilePath = filepath.Join(outDir, ViewsFolder, "slo-"+slo.Metadata.Name+".tf")
		return FileFromTmpl(NameViewTmpl, tfFilePath, vc)
	case NameDashboardTmpl:
		dc, err := DashConfigFromSLO(slo)
		if err != nil {
			return err
		}
		tfFilePath = filepath.Join(outDir, DashboardsFolder, "slo-"+slo.Metadata.Name+".tf")
		return FileFromTmpl(NameDashboardTmpl, tfFilePath, dc)
	case NameMonitorTmpl:
		mc, err := MonitorConfigFromOpenSLO(slo)
		if err != nil {
			return err
		}
		tfFilePath = filepath.Join(outDir, MonitorsFolder, "slo-"+slo.Metadata.Name+".tf")
		return FileFromTmpl(NameMonitorTmpl, tfFilePath, mc)
	}

	return nil
}

const (
	VarNameMonRootFolder  = "slo_mon_root_folder_id"
	VarNameDashRootFolder = "slo_dash_root_folder_id"
)

type TFModules struct {
	Path string
	Vars []string
}

func SetupOutDir(c GenConf) error {

	var err error
	err = EnsureDir(c.OutDir, false)

	if err != nil {
		return err
	}

	mainPath := filepath.Join(c.OutDir, "main.tf")

	err = FileFromTmpl(NameMainTmpl, mainPath, c)
	if err != nil {
		return err
	}

	modules := []TFModules{
		{Path: filepath.Join(c.OutDir, ViewsFolder), Vars: nil},
		{Path: filepath.Join(c.OutDir, MonitorsFolder), Vars: []string{VarNameMonRootFolder}},
		{Path: filepath.Join(c.OutDir, DashboardsFolder), Vars: []string{VarNameDashRootFolder}},
	}

	for _, p := range modules {
		err := EnsureDir(p.Path, c.Clean)
		if err != nil {
			return err
		}

		moduleTFPath := filepath.Join(p.Path, "module_interface.tf")
		FileFromTmpl(NameModuleTmpl, moduleTFPath, p.Vars)
		if err != nil {
			return err
		}
	}

	return nil
}

func FileFromTmpl(name string, path string, data interface{}) error {
	moduleTmpl := tfTemplates.Lookup(name)
	buff := &bytes.Buffer{}
	err := moduleTmpl.Execute(buff, data)

	if err != nil {
		return err
	}
	return os.WriteFile(path, buff.Bytes(), 0755)
}

func GenFoldersTF(srvList []string, outDir string) error {
	path := filepath.Join(outDir, MonitorsFolder, "folders.tf")
	err := FileFromTmpl(NameMonitorFolderTmpl, path, srvList)

	if err != nil {
		return err
	}

	path = filepath.Join(outDir, DashboardsFolder, "folders.tf")
	err = FileFromTmpl(NameDashFolderTmpl, path, srvList)

	return err
}
