package libs

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
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

//go:embed templates/**
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
}

func init() {
	var err error

	tfTemplates, err = template.ParseFS(tmplFiles, "templates/terra/**")

	if err != nil {
		panic(err)
	}
}

const ViewPrefix = "slogen_tf"

// GenTerraform for all openslo files in the path
func GenTerraform(slos map[string]*SLO, c GenConf) (string, error) {

	err := SetupOutDir(c)
	if err != nil {
		BadResult("error setting up path : %s", err)
		return "", err
	}

	srvMap := map[string]bool{}

	for path, s := range slos {
		s.ViewName = GiveScheduleViewName(*s)
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

	err = GenFoldersTF(srvList, c.OutDir)
	if err != nil {
		return "", err
	}

	err = GenOverviewTF(slos, c)

	return "", err
}

func ExecSLOTmpl(tmplName string, slo SLO, outDir string) error {

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
