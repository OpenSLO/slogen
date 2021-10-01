package libs

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func giveOverviewListQuery(dashVars []string) string {

	clauses := []string{}
	for _, v := range dashVars {
		clauses = append(clauses, fmt.Sprintf("(\"{{%s}}\"=\"*\" or %s=\"{{%s}}\")", v, v, v))
	}

	wherePart := strings.Join(clauses, " and ")

	query := "_view=slogen_tfm_* | where " + wherePart + `
| sum(sliceGoodCount) as GoodReqs, sum(sliceTotalCount) as TotalReqs by Service,SLOName
| (GoodReqs/TotalReqs)*100 as SLAVal
| order by SLAVal asc
| format("%.2f%%",SLAVal)  as SLA | fields  Service, SLOName, SLA, GoodReqs, TotalReqs
`

	return query
}

type SLOOverviewDashConf struct {
	Query    string
	DashVars []string
}

func GenOverviewTF(s map[string]*SLO, c GenConf) error {

	dashVars := giveMostCommonVars(s, 3)
	query := giveOverviewListQuery(dashVars)

	dashVars = append([]string{"Service"}, dashVars...)
	conf := SLOOverviewDashConf{
		Query:    query,
		DashVars: dashVars,
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

	varList := []string{}
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
