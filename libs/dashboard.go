package libs

import (
	"bytes"
	_ "embed"
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
	target := *(sloConf.Spec.Objectives[0].BudgetTarget)

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

	return panels, nil
}

type gaugeVizSettingParams struct {
	TargetMidBad float64
	TargetBad    float64
}

func giveSLOGaugePanels(s SLO) ([]SearchPanel, error) {

	target := *(s.Spec.Objectives[0].BudgetTarget)
	tmpl, err := template.New("vizGaugeSetting").Delims("[[", "]]").Parse(vizSettingGauge)

	if err != nil {
		return nil, err
	}

	vizParams := gaugeVizSettingParams{
		TargetMidBad: (1 + target) * 50,
		TargetBad:    (target) * 100,
	}
	vizBuff := bytes.Buffer{}

	err = tmpl.Execute(&vizBuff, vizParams)

	if err != nil {
		return nil, err
	}

	query := givePanelQuery(KeyGaugeToday, s.ViewName)

	vizSetting := vizBuff.String()
	today := SearchPanel{
		Key:            KeyGaugeToday,
		Title:          "Today's Availability",
		Desc:           "#good-requests / #requests since start of day",
		VisualSettings: vizSetting,
		Query:          query,
		TimeRange:      "today",
	}
	week := SearchPanel{
		Key:            KeyGaugeWeek,
		Title:          "Week's Availability",
		Desc:           "#good-requests / #requests since start of week",
		VisualSettings: vizSetting,
		Query:          query,
		TimeRange:      "week",
	}
	month := SearchPanel{
		Key:            KeyGaugeMonth,
		Title:          "Month's Availability",
		Desc:           "#good-requests / #requests since start of month",
		VisualSettings: vizSetting,
		Query:          query,
		TimeRange:      "month",
	}

	panels := []SearchPanel{today, week, month}

	return panels, nil
}

func giveHourlyBurnRatePanel(s SLO) (SearchPanel, error) {
	query := givePanelQuery(KeyPanelHourlyBurn, s.ViewName)
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
	query := givePanelQuery(KeyPanelBurnTrend, s.ViewName)
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
	query := givePanelQuery(KeyPanelBudgetLeft, s.ViewName)
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

type PanelKey string

const (
	KeyGaugeToday      PanelKey = "gauge-today"
	KeyGaugeWeek       PanelKey = "gauge-week"
	KeyGaugeMonth      PanelKey = "gauge-month"
	KeyPanelHourlyBurn PanelKey = "hourly-burn-rate"
	KeyPanelBurnTrend  PanelKey = "burn-trend"
	KeyPanelBudgetLeft PanelKey = "budget-left"
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
}

//go:embed templates/visual-settings/gauge-viz-settings.gojson
var vizSettingGauge string

//go:embed templates/visual-settings/burn-trend.gojson
var vizSettingBurnTrend string

//go:embed templates/visual-settings/budget-left.gojson
var vizSettingBudgetLeft string

//go:embed templates/visual-settings/hourly-burn-rate.gojson
var vizSettingHourlyBurn string

func givePanelQuery(Key PanelKey, view string) string {

	var qPart string
	switch Key {
	case KeyGaugeToday, KeyGaugeWeek, KeyGaugeMonth:
		qPart = gaugeQueryPart
	case KeyPanelHourlyBurn:
		qPart = hourlyBurnQueryPart
	case KeyPanelBurnTrend:
		qPart = burnTrendQueryPart
	case KeyPanelBudgetLeft:
		qPart = budgetLeftQueryPart
	default:
		return ""
	}

	return "_view=" + view + qPart
}
