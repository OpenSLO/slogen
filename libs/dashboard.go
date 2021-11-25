package libs

import (
	"bytes"
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"text/template"
)

type SLODashboard struct {
	Service       string
	SLOName       string
	Title         string
	Desc          string
	Theme         string
	Target        float64
	RefreshFreq   int
	SearchPanels  []SearchPanel
	Layout        []LayoutItem
	Viz           VisualSetting
	StrYAMLConfig string
	Labels        map[string]string
	Fields        []string
	ViewName      string
}

type SearchPanel struct {
	Key                 PanelKey
	Title               string
	Desc                string
	VisualSettings      string
	Query               string
	TimeRange           string
	IsRelativeTimeRange bool
}

type VisualSetting struct {
	Mode         string
	Type         string
	XAxisUnit    string
	XAxisTitle   string
	MarkerType   string
	DisplayType  string
	LineDashType string
}

type LayoutItem struct {
	Key       PanelKey
	Structure string
}

func DashConfigFromSLO(sloConf SLO) (*SLODashboard, error) {
	sloName := sloConf.Metadata.Name
	target := sloConf.Target()

	configYamlBytes, err := yaml.Marshal(sloConf)

	if err != nil {
		return nil, err
	}

	panels, err := giveSLOPanels(sloConf)

	if err != nil {
		return nil, err
	}

	conf := &SLODashboard{
		Service:       sloConf.Spec.Service,
		SLOName:       sloName,
		Title:         sloConf.Metadata.DisplayName,
		Desc:          sloConf.Spec.Description,
		Theme:         "Light",
		Target:        target * 100,
		RefreshFreq:   300,
		Layout:        sloLayout,
		SearchPanels:  panels,
		StrYAMLConfig: string(configYamlBytes),
		Labels:        sloConf.Labels,
		Fields:        giveMapKeys(sloConf.Fields),
		ViewName:      sloConf.ViewName,
	}

	return conf, nil
}

func giveSLOPanels(s SLO) ([]SearchPanel, error) {
	var panels []SearchPanel
	gauge, err := giveSLOGaugePanels(s)
	if err != nil {
		return nil, err
	}

	burnPanel, err := giveHourlyBurnRatePanel(s)

	if err != nil {
		return nil, err
	}

	trendPanel, err := giveTrendOfBurnRatePanel(s)

	if err != nil {
		return nil, err
	}

	budgetPanel, err := giveBudgetDepletionPanel(s)
	if err != nil {
		return nil, err
	}

	panels = append(gauge, burnPanel, trendPanel, budgetPanel)

	if len(s.Fields) > 0 {
		breakdownPanel, err := giveBreakdownPanel(s)
		if err != nil {
			return nil, err
		}
		panels = append(panels, breakdownPanel)
	}

	return panels, nil
}

type gaugeVizSettingParams struct {
	TargetMidBad float64
	TargetBad    float64
}

func giveSLOGaugePanels(s SLO) ([]SearchPanel, error) {

	target := *(s.Spec.Objectives[0].BudgetTarget)

	vizParams := gaugeVizSettingParams{
		TargetMidBad: (1 + target) * 50,
		TargetBad:    (target) * 100,
	}

	vizSettingStr, err := GiveStrFromTmpl(vizSettingGauge, vizParams)

	if err != nil {
		return nil, err
	}

	query, err := givePanelQuery(s, KeyGaugeToday)

	if err != nil {
		return nil, err
	}

	today := SearchPanel{
		Key:            KeyGaugeToday,
		Title:          "Today's Availability",
		Desc:           "#good-requests / #requests since start of day",
		VisualSettings: vizSettingStr,
		Query:          query,
		TimeRange:      "today",
	}
	week := SearchPanel{
		Key:            KeyGaugeWeek,
		Title:          "Week's Availability",
		Desc:           "#good-requests / #requests since start of week",
		VisualSettings: vizSettingStr,
		Query:          query,
		TimeRange:      "week",
	}
	month := SearchPanel{
		Key:            KeyGaugeMonth,
		Title:          "Month's Availability",
		Desc:           "#good-requests / #requests since start of month",
		VisualSettings: vizSettingStr,
		Query:          query,
		TimeRange:      "month",
	}

	panels := []SearchPanel{today, week, month}

	return panels, nil
}

func giveHourlyBurnRatePanel(s SLO) (SearchPanel, error) {
	query, err := givePanelQuery(s, KeyPanelHourlyBurn)

	if err != nil {
		return SearchPanel{}, err
	}

	panel := SearchPanel{
		Key:                 KeyPanelHourlyBurn,
		Title:               "Hourly Burn Rate",
		Desc:                "(ErrorsObserved)/(ErrorBudget) for the hour buckets where ErrorBudget = (1-SLO)*TotalRequests",
		VisualSettings:      vizSettingHourlyBurn,
		Query:               query,
		TimeRange:           "-24h",
		IsRelativeTimeRange: true,
	}
	return panel, nil
}

func giveTrendOfBurnRatePanel(s SLO) (SearchPanel, error) {

	query, err := givePanelQuery(s, KeyPanelBurnTrend)

	if err != nil {
		return SearchPanel{}, err
	}

	panel := SearchPanel{
		Key:            KeyPanelBurnTrend,
		Title:          "Burn rate trend compared to last 7 days  (upto current time of the day)",
		Desc:           "Today's burn rate (so far) along with last 7 days (till the same time as today)",
		VisualSettings: vizSettingBurnTrend,
		Query:          query,
		TimeRange:      "today",
	}
	return panel, nil
}

func giveBudgetDepletionPanel(s SLO) (SearchPanel, error) {
	query, err := givePanelQuery(s, KeyPanelBudgetLeft)

	if err != nil {
		return SearchPanel{}, err
	}

	panel := SearchPanel{
		Key:            KeyPanelBudgetLeft,
		Title:          "Budget remaining",
		Desc:           "Error budget from start of month",
		VisualSettings: vizSettingBudgetLeft,
		Query:          query,
		TimeRange:      "month",
	}
	return panel, nil
}

func giveBreakdownPanel(s SLO) (SearchPanel, error) {
	query, err := givePanelQuery(s, KeyPanelBreakDown)

	panel := SearchPanel{
		Key:            KeyPanelBreakDown,
		Title:          "SLO Breakdown",
		Desc:           "reliability stats by fields specified in the config",
		VisualSettings: vizSettingBreakdownPanel,
		Query:          query,
		TimeRange:      "month",
	}

	return panel, err
}

type PanelKey string

const (
	KeyGaugeToday      PanelKey = "gauge-today"
	KeyGaugeWeek       PanelKey = "gauge-week"
	KeyGaugeMonth      PanelKey = "gauge-month"
	KeyPanelHourlyBurn PanelKey = "hourly-burn-rate"
	KeyPanelBurnTrend  PanelKey = "burn-trend"
	KeyPanelBudgetLeft PanelKey = "budget-left"
	KeyPanelBreakDown  PanelKey = "breakdown"
)

var sloLayout = []LayoutItem{
	{
		Key:       KeyGaugeToday,
		Structure: `{\"height\":6,\"width\":6,\"x\":0,\"y\":0}`,
	},
	{
		Key:       KeyGaugeWeek,
		Structure: `{\"height\":6,\"width\":6,\"x\":0,\"y\":6}`,
	},
	{
		Key:       KeyGaugeMonth,
		Structure: `{\"height\":6,\"width\":6,\"x\":0,\"y\":12}`,
	},
	{
		Key:       KeyPanelHourlyBurn,
		Structure: `{\"height\":6,\"width\":18,\"x\":6,\"y\":0}`,
	},
	{
		Key:       KeyPanelBurnTrend,
		Structure: `{\"height\":6,\"width\":18,\"x\":6,\"y\":6}`,
	},
	{
		Key:       KeyPanelBudgetLeft,
		Structure: `{\"height\":6,\"width\":18,\"x\":6,\"y\":12}`,
	},
	{
		Key:       "text-panel-config",
		Structure: `{\"height\":6,\"width\":12,\"x\":0,\"y\":18}`,
	},
	{
		Key:       "text-panel-details",
		Structure: `{\"height\":6,\"width\":12,\"x\":12,\"y\":18}`,
	},
	{
		Key:       KeyPanelBreakDown,
		Structure: `{\"height\":10,\"width\":24,\"x\":0,\"y\":24}`,
	},
}

//go:embed templates/visual-settings/gauge-viz-settings.gojson
var vizSettingGauge string

//go:embed templates/visual-settings/burn-trend.gojson
var vizSettingBurnTrend string

//go:embed templates/visual-settings/budget-left.gojson
var vizSettingBudgetLeft string

//go:embed templates/visual-settings/hourly-burn-rate.gojson
var vizSettingHourlyBurn string

//go:embed templates/visual-settings/breakdown-panel.gojson
var vizSettingBreakdownPanel string


//go:embed templates/visual-settings/forecasted-panel.json
var vizSettingBudgetForecastPanel string

func givePanelQuery(s SLO, key PanelKey) (string, error) {
	//queryStr := givePanelQueryStr(key, s.ViewName)
	queryTmplStr := ""
	if s.Spec.BudgetingMethod == BudgetingMethodNameTimeSlices {
		queryTmplStr = givePanelQueryTimesliceStr(key, s.ViewName)
	} else {
		queryTmplStr = givePanelQueryStr(key, s.ViewName)
	}

	wherePart := giveWhereClause(giveMapKeys(s.Fields))

	tmplParams := struct {
		Target               float64
		TimesliceRatioTarget float64
		GroupByStr string
	}{
		Target:               s.Target(),
		TimesliceRatioTarget: s.TimesliceTarget(),
		GroupByStr : giveFieldsGroupByStr(s.Fields),
	}

	queryPart, err := GiveStrFromTmpl(queryTmplStr, tmplParams)
	return fmt.Sprintf("_view=%s %s %s", s.ViewName, wherePart, queryPart), err

}

func givePanelQueryStr(Key PanelKey, view string) string {

	var qPart string
	switch Key {
	case KeyGaugeToday, KeyGaugeWeek, KeyGaugeMonth:
		qPart = gaugeQueryPartForOccurrences
	case KeyPanelHourlyBurn:
		qPart = hourlyBurnQueryPartForOccurrences
	case KeyPanelBurnTrend:
		qPart = burnTrendQueryPartForOccurrences
	case KeyPanelBudgetLeft:
		qPart = budgetLeftQueryPart
	case KeyPanelBreakDown:
		qPart = breakDownPanelQueryOccurrences
	default:
		return ""
	}

	return qPart
}

func givePanelQueryTimesliceStr(Key PanelKey, view string) string {

	var qPart string
	switch Key {
	case KeyGaugeToday, KeyGaugeWeek, KeyGaugeMonth:
		qPart = gaugeQueryPartForTimeslice
	case KeyPanelHourlyBurn:
		qPart = hourlyBurnQueryPartForTimeslice
	case KeyPanelBurnTrend:
		qPart = burnTrendQueryPartForTimeslice
	case KeyPanelBudgetLeft:
		qPart = budgetLeftQueryTimeSlicesPart
	case KeyPanelBreakDown:
		qPart = breakDownPanelQueryTimeslices
	default:
		return ""
	}

	return qPart
}

func GiveStrFromTmpl(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("giveQueryFromTmpl").Parse(tmplStr)

	if err != nil {
		return "", err
	}

	buff := bytes.Buffer{}

	err = tmpl.Execute(&buff, data)

	return buff.String(), err
}
