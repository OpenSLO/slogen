package libs

import (
	"bytes"
	"text/template"
)

type SLOMonitorConfig struct {
	Name       string
	Desc       string
	Service    string
	Objectives []SLOObjective
	//Notify []
}

const (
	TriggerNameCritical        = "Critical"
	TriggerNameResolveCritical = "ResolvedCritical"
	TriggerNameWarning         = "Warning"
	TriggerNameResolveWarning  = "ResolvedWarning"
)

func giveDefaultTriggers() []string {
	return []string{TriggerNameCritical, TriggerNameResolveCritical}
}

type Notification struct {
	TriggerFor     []string `yaml:"triggerFor,omitempty"`
	ConnectionType string   `yaml:"connectionType"`
	ConnectionID   string   `yaml:"connectionID,omitempty"`
	Subject        string   `yaml:"subject,omitempty"`
	Recipients     []string `yaml:"recipients,omitempty"`
	MessageBody    string   `yaml:"messageBody,omitempty"`
	TimeZone       string   `yaml:"timeZone,omitempty"`
}

type SLOObjective struct {
	Suffix        string
	Query         string
	Field         string
	TimeRange     string
	ValueWarning  float64
	ValueCritical float64
	Notifications []Notification `yaml:"notifications,omitempty"`
}

// BurnRate only supports 2 window with first one having
type BurnRate struct {
	View          string         `yaml:"view,omitempty"`
	Budget        float64        `yaml:"budget,omitempty"`
	ShortWindow   string         `yaml:"shortWindow"`
	ShortLimit    float64        `yaml:"shortLimit"`
	LongWindow    string         `yaml:"longWindow"`
	LongLimit     float64        `yaml:"longLimit"`
	Notifications []Notification `yaml:"notifications,omitempty"`
}

const MultiWindowMultiBurnTmpl = `_view={{.View}}
| timeslice {{.ShortWindow}} 
| sum(sliceGoodCount) as tmGood, sum(sliceTotalCount) as tmCount  group by _timeslice
| tmGood/tmCount as tmSLO 
| (tmCount-tmGood) as tmBad 
| total tmCount as totalCount  
| totalCount*(1-{{.Budget}}) as errorBudget
| ((tmBad/tmCount)/(1-{{.Budget}})) as sliceBurnRate
| if(queryEndTime() - _timeslice <= {{.ShortWindow}},sliceBurnRate, 0  )  as latestBurnRate 
| sum(tmGood) as totalGood, max(totalCount) as totalCount, max(latestBurnRate) as latestBurnRate 
| (1-(totalGood/totalCount))/(1-{{.Budget}}) as longBurnRate
| if (longBurnRate > {{.LongLimit}} , 1,0) as long_burn_exceeded
| if ( latestBurnRate > {{.ShortLimit}}, 1,0) as short_burn_exceeded
| long_burn_exceeded + short_burn_exceeded as combined_burn
`

func MonitorConfigFromOpenSLO(sloConf SLO) (*SLOMonitorConfig, error) {

	tmpl, err := template.New("monitor").Parse(MultiWindowMultiBurnTmpl)

	if err != nil {
		return nil, err
	}

	mConf := &SLOMonitorConfig{
		Name:    sloConf.ObjectHeader.Metadata.Name,
		Desc:    sloConf.Spec.Description,
		Service: sloConf.Spec.Service,
	}

	var objectives []SLOObjective

	alerts := append(sloConf.BurnRateAlerts, sloConf.Alerts.BurnRate...)

	for _, alert := range alerts {

		alert.View = sloConf.ViewName
		buf := bytes.Buffer{}
		err = tmpl.Execute(&buf, alert)
		if err != nil {
			return nil, err
		}
		obj := SLOObjective{
			Suffix:        alert.ShortWindow + "_" + alert.LongWindow,
			Query:         buf.String(),
			Field:         "combined_burn",
			TimeRange:     alert.LongWindow,
			ValueWarning:  1,
			ValueCritical: 2,
			Notifications: alert.Notifications,
		}
		objectives = append(objectives, obj)
	}
	mConf.Objectives = objectives
	return mConf, nil
}

//func GiveBurnMonitorConf(sloConf SLO) (*SLOMonitorConfig, error) {
//
//	tmpl, err := template.New("monitor").Parse(GoodByTotalQueryTmpl)
//
//	if err != nil {
//		return nil, err
//	}
//
//	buf := bytes.Buffer{}
//	err = tmpl.Execute(&buf, sloConf.Spec.Objectives[0].RatioMetrics)
//	if err != nil {
//		return nil, err
//	}
//	m := &SLOMonitorConfig{
//		Name:    sloConf.ObjectHeader.Metadata.Name,
//		Desc:    sloConf.Spec.Description,
//		Service: sloConf.Spec.Service,
//		Objectives: []SLOObjective{
//			{
//				Field:         "SLO",
//				Query:         buf.String(),
//				TimeRange:     "24h",
//				ValueWarning:  (*sloConf.Spec.Objectives[0].BudgetTarget + 1.0) / 2,
//				ValueCritical: *sloConf.Spec.Objectives[0].BudgetTarget,
//			},
//		},
//	}
//	return m, err
//}
