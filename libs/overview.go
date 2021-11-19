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

    wherePart := "| where " +  strings.Join(clauses, " and ")
   
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

// giveMostCommonVars top n most common label or fields found
func giveMostCommonVars(slos SLOMap, n int) []string {

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
