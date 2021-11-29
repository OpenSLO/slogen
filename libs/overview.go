package libs

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func giveOverviewListQuery(dashVars []string) string {

	wherePart := giveWhereClause(dashVars)

	query := "_view=slogen_tf_* " + wherePart + `
| sum(sliceGoodCount) as GoodCount, sum(sliceTotalCount) as TotalCount by Service, SLOName
| (GoodCount/TotalCount)*100 as SLAVal
| order by SLAVal asc
| SLOName as ObjectiveName
| format("%.2f%%",SLAVal)  as Availability 
| fields  Service, ObjectiveName, Availability, GoodCount, TotalCount
`

	return query
}

func giveWhereClause(dashVars []string) string {

	if len(dashVars) == 0 {
		return ""
	}

	var clauses []string
	for _, v := range dashVars {
		clauses = append(clauses, fmt.Sprintf("(\"{{%s}}\"=\"*\" or %s=\"{{%s}}\")", v, v, v))
	}

	wherePart := "| where " + strings.Join(clauses, " and ")

	return wherePart
}

func giveOverviewWeeksQuery(dashVars []string) string {

	var clauses []string
	for _, v := range dashVars {
		clauses = append(clauses, fmt.Sprintf("(\"{{%s}}\"=\"*\" or %s=\"{{%s}}\")", v, v, v))
	}

	wherePart := strings.Join(clauses, " and ")

	query := "_view=slogen_tf_* | where " + wherePart + `
| timeslice 1d
| sum(sliceGoodCount) as GoodReqs, sum(sliceTotalCount) as TotalReqs by _timeslice,Service, SLOName
| (GoodReqs/TotalReqs) as SLAVal
| avg(SLAVal)  as AvgAvailability by _timeslice,Service
| transpose row _timeslice column Service
`

	return query
}

type SLOOverviewDashConf struct {
	QueryTable string
	QueryDaily string
	DashVars   []string
}

func GenOverviewTF(s map[string]*SLO, c GenConf) error {
	err := GenGlobalOverviewTF(s, c)
	if err != nil {
		return err
	}

	err = GenServiceOverviewDashboard(s, c.OutDir)

	return err
}

func GenGlobalOverviewTF(s map[string]*SLO, c GenConf) error {

	dashVars := giveMostCommonVars(s, 3)
	query := giveOverviewListQuery(dashVars)

	dashVars = append([]string{"Service"}, dashVars...)
	conf := SLOOverviewDashConf{
		QueryTable: query,
		QueryDaily: giveOverviewWeeksQuery(dashVars),
		DashVars:   dashVars,
	}
	path := filepath.Join(c.OutDir, DashboardsFolder, "overview.tf")
	return FileFromTmpl(NameGlobalTrackerTmpl, path, conf)
}

type SLOMap map[string]*SLO

func giveMostCommonVarsFromSLOSLice(slos []SLO, n int) []string {
	vCount := map[string]int{}

	for _, s := range slos {
		for k := range s.Fields {
			vCount[k] = vCount[k] + 1
		}

		for k := range s.Labels {
			vCount[k] = vCount[k] + 1
		}
	}

	var varList []string
	for k := range vCount {
		varList = append(varList, k)
	}

	if len(varList) <= n {
		return varList
	}

	sort.Slice(varList, func(i, j int) bool {
		ki := varList[i]
		kj := varList[j]
		return vCount[ki] > vCount[kj]
	})

	return varList[:n]
}

// giveMostCommonVars top n most common label or fields found
func giveMostCommonVars(slos SLOMap, n int) []string {

	slc := giveSLOMapToSlice(slos)
	return giveMostCommonVarsFromSLOSLice(slc, n)
}

func giveSLOMapToSlice(s SLOMap) []SLO {
	var slc []SLO
	for _, slo := range s {
		slc = append(slc, *slo)
	}

	return slc
}

func giveServiceToSLOMap(slos map[string]*SLO) map[string][]SLO {
	srvMap := map[string][]SLO{}

	for _, s := range slos {
		srvMap[s.Spec.Service] = append(srvMap[s.Spec.Service], *s)
	}

	return srvMap
}

func GenServiceOverviewDashboard(sloPathMap map[string]*SLO, outDir string) error {
	srvMap := giveServiceToSLOMap(sloPathMap)

	for srv, slos := range srvMap {
		rows, err := getOverviewRows(slos)
		if err != nil {
			return err
		}

		layout := getOverviewLayout(rows)

		conf := ServiceOverviewDashboard{
			Service: srv,
			Rows:    rows,
			Layout:  layout,
			Vars: giveMostCommonVarsFromSLOSLice(slos,4),
		}

		path := filepath.Join(outDir, DashboardsFolder, "overview-" +srv +".tf")
		err = FileFromTmpl(NameServiceTrackerTmpl, path, conf)
		if err != nil {
			return err
		}
	}
	return nil
}

type ServiceOverviewDashboard struct {
	Service string
	Rows    []ServiceOverviewRow
	Layout  []LayoutItem
	Vars    []string
}

type ServiceOverviewRow struct {
	SLOName string
	Panels  []SearchPanel
	SLOConf SLO
}

func getOverviewRows(s []SLO) ([]ServiceOverviewRow, error) {

	var rows []ServiceOverviewRow
	for _, slo := range s {
		r, err := getOverviewRow(slo)

		if err != nil {
			return nil, err
		}
		rows = append(rows, r)
	}

	return rows, nil
}

func getOverviewRow(s SLO) (ServiceOverviewRow, error) {
	name := s.Metadata.Name

	row := ServiceOverviewRow{
		SLOName: name,
		Panels:  nil,
		SLOConf: s,
	}

	panels, err := giveSLOGaugePanels(s)
	if err != nil {
		return row, err
	}

	for i, p := range panels {
		panels[i].Key = PanelKey(name + "-" + string(p.Key))
	}

	row.Panels = panels

	return row, err
}

func getOverviewLayout(rows []ServiceOverviewRow) []LayoutItem {
	var layout []LayoutItem

	h, w, x, y := 6, 6, 0, 0
	for _, r := range rows {

		layout = append(layout, LayoutItem{
			Key:       r.SLOName + "-text-overview",
			Structure: fmt.Sprintf(`{\"height\":%d,\"width\":%d,\"x\":%d,\"y\":%d}`, h, w, x, y),
		})

		for _, p := range r.Panels {
			x = x + w
			layout = append(layout, LayoutItem{
				Key:       string(p.Key),
				Structure: fmt.Sprintf(`{\"height\":%d,\"width\":%d,\"x\":%d,\"y\":%d}`, h, w, x, y),
			})
		}

		x = 0
		y = y + h
	}

	return layout
}
