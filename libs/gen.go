package libs

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/OpenSLO/slogen/libs/specs"
	"github.com/OpenSLO/slogen/libs/sumologic"
	oslo "github.com/agaurav/oslo/pkg/manifest/v1"
	"os"
	"path/filepath"
	"sort"
	"text/template"
)

var tfTemplates *template.Template

const NameDashboardTmpl = "dashboard.tf.gotf"
const NameViewTmpl = "sched-view.tf.gotf"
const NameMonitorTmpl = "monitor.tf.gotf"
const NameModuleTmpl = "module_interface.tf.gotf"
const NameDashFolderTmpl = "dash-folders.tf.gotf"
const NameMonitorFolderTmpl = "monitor-folders.tf.gotf"
const NameSLOFolderTmpl = "slo-folders.tf.gotf"
const NameGlobalTrackerTmpl = "global-tracker.tf.gotf"
const NameServiceTrackerTmpl = "service-overview.tf.gotf"

//go:embed templates/terraform/**/*.tf.gotf
var tmplFiles embed.FS

//go:embed templates/terraform/main.tf.gotf
var tmplMainTFStr string

var alertPolicyMap = map[string]oslo.AlertPolicy{}
var notificationTargetMap = map[string]oslo.AlertNotificationTarget{}

const (
	BuildFolder      = "build"
	ViewsFolder      = "views"
	MonitorsFolder   = "monitors"
	DashboardsFolder = "dashboards"
	NativeSLOFolder  = "slos"
)

type GenConf struct {
	OutDir               string
	DashFolder           string
	MonitorFolder        string
	ViewPrefix           string
	SLORootFolder        string
	SLOMonitorRootFolder string
	DoPlan               bool
	DoApply              bool
	IgnoreError          bool
	Clean                bool
	AsModule             bool
	UseViewHash          bool
	OnlyNative           bool
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

	slosAlpha, _ := splitMultiVerse(slosMv)
	err := SetupOutDir(c)
	if err != nil {
		BadResult("error setting up path : %s", err)
		return "", err
	}

	//pretty.Println(slosV1)

	err = genTerraformForV1(slosMv, c)

	if err != nil {
		panic(err)
	}

	//to be moved to conditional compilation later
	//graph.MakeOSLODepGraph(slosMv, c)

	return genTerraformForAlpha(slosAlpha, c)
}

func genTerraformForV1(slos map[string]*SLOMultiVerse, c GenConf) error {

	v1Path := filepath.Join(c.OutDir, NativeSLOFolder)

	fillAlertPolicyMap(slos)
	fillNotificationTargetMap(slos)

	srvMap := map[string]bool{}

	var err error
	var sloStr, monitorsStr string

	for _, sloM := range slos {
		if sloM.SLO != nil {
			slo := sloM.SLO

			// handle sumologic specific stuff
			if sumologic.IsSource(*slo) {
				sloStr, monitorsStr, err = sumologic.GiveTerraform(alertPolicyMap, notificationTargetMap, *slo)
			}

			err = os.WriteFile(filepath.Join(v1Path, fmt.Sprintf("slo_%s.tf", slo.Metadata.Name)), []byte(sloStr), 0755)
			if err != nil {
				return err
			}
			srvMap[slo.Spec.Service] = true

			if monitorsStr != "" {
				err = os.WriteFile(filepath.Join(v1Path, fmt.Sprintf("slo_monitors_%s.tf", slo.Metadata.Name)), []byte(monitorsStr), 0755)
			}
		}
	}

	srvList := GiveKeys(srvMap)
	sort.Strings(srvList)

	GenSLOFoldersTF(srvList, c)

	return nil
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

	if !c.OnlyNative {
		err = GenFoldersTF(srvList, c)
		if err != nil {
			return "", err
		}

		err = GenOverviewTF(slosAlpha, c)
	}

	return "", err
}

func fillAlertPolicyMap(slos map[string]*SLOMultiVerse) {
	for _, slo := range slos {
		ap := slo.AlertPolicy
		if ap != nil {
			alertPolicyMap[ap.Metadata.Name] = *ap
		}
	}
}

func fillNotificationTargetMap(slos map[string]*SLOMultiVerse) {
	for _, slo := range slos {
		nt := slo.AlertNotificationTarget
		if nt != nil {
			notificationTargetMap[nt.Metadata.Name] = *nt
		}
	}
}

func splitMultiVerse(slos map[string]*SLOMultiVerse) (map[string]*SLOv1Alpha, map[string]*specs.OpenSLOSpec) {

	alpha := map[string]*SLOv1Alpha{}
	v1 := map[string]*specs.OpenSLOSpec{}

	for path, sloMv := range slos {
		if sloMv.SLOAlpha != nil {
			alpha[path] = sloMv.SLOAlpha
		}
		if sloMv.SLO != nil {
			v1[path] = sloMv.SLO
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
	VarNameMonRootFolder        = "slo_mon_root_folder_id"
	VarNameDashRootFolder       = "slo_dash_root_folder_id"
	VarNameNativeSLORootFolder  = "slo_root_folder_id"
	VarNameSLOMonitorRootFolder = "slo_monitor_root_folder_id"
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

	err = GenMainTF(mainPath, c)
	if err != nil {
		return err
	}

	var modules []TFModules

	if c.OnlyNative {
		modules = []TFModules{
			{Path: filepath.Join(c.OutDir, NativeSLOFolder), Vars: []string{VarNameNativeSLORootFolder, VarNameSLOMonitorRootFolder}},
		}
	} else {
		modules = []TFModules{
			{Path: filepath.Join(c.OutDir, ViewsFolder), Vars: nil},
			{Path: filepath.Join(c.OutDir, MonitorsFolder), Vars: []string{VarNameMonRootFolder}},
			{Path: filepath.Join(c.OutDir, DashboardsFolder), Vars: []string{VarNameDashRootFolder}},
			{Path: filepath.Join(c.OutDir, NativeSLOFolder), Vars: []string{VarNameNativeSLORootFolder, VarNameSLOMonitorRootFolder}},
		}
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

func GenMainTF(path string, conf GenConf) error {
	mainTFTmpl := template.Must(template.New("main.tf").Parse(tmplMainTFStr))

	buff := &bytes.Buffer{}

	err := mainTFTmpl.Execute(buff, conf)

	if err != nil {
		return err
	}

	return os.WriteFile(path, buff.Bytes(), 0644)
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

func GenFoldersTF(srvList []string, conf GenConf) error {

	outDir := conf.OutDir

	fmt.Println(srvList, "srvList")

	if !conf.OnlyNative {
		path := filepath.Join(outDir, MonitorsFolder, "folders.tf")
		err := FileFromTmpl(NameMonitorFolderTmpl, path, srvList)

		if err != nil {
			return err
		}

		path = filepath.Join(outDir, DashboardsFolder, "folders.tf")
		err = FileFromTmpl(NameDashFolderTmpl, path, srvList)

		if err != nil {
			return err
		}
	}

	return nil
}

func GenSLOFoldersTF(srvList []string, conf GenConf) error {

	outDir := conf.OutDir

	path := filepath.Join(outDir, NativeSLOFolder, "folders.tf")
	err := FileFromTmpl(NameSLOFolderTmpl, path, srvList)

	return err
}
