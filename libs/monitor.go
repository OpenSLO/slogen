package libs

import (
	"bytes"
	"sort"
	"text/template"
	"time"
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
	ShortWindow   string         `yaml:"shortWindow"`
	ShortLimit    float64        `yaml:"shortLimit"`
	LongWindow    string         `yaml:"longWindow"`
	LongLimit     float64        `yaml:"longLimit"`
	Notifications []Notification `yaml:"notifications,omitempty"`
}

const MultiWindowMultiBurnTmpl = `_view={{.View}}
| timeslice {{.ShortWindow}} 
| sum(sliceGoodCount) as tmGood, sum(sliceTotalCount) as tmCount  group by _timeslice
| fillmissing timeslice(1m)
| tmGood/tmCount as tmSLO 
| (tmCount-tmGood) as tmBad 
| total tmCount as totalCount
| totalCount*(1-{{.Target}}) as errorBudget
| if(tmCount>0,tmBad/tmCount,0) as errorRate
| (errorRate/(1-{{.Target}})) as sliceBurnRate
| if(queryEndTime() - _timeslice <= {{.ShortWindow}},sliceBurnRate, 0  )  as latestBurnRate 
| sum(tmGood) as totalGood, max(totalCount) as totalCount, max(latestBurnRate) as latestBurnRate 
| if(totalCount>0,totalGood/totalCount,0) as longErrorRate
| (1-longErrorRate)/(1-{{.Target}}) as longBurnRate
| if (longBurnRate > {{.LongLimit}} , 1,0) as long_burn_exceeded
| if ( latestBurnRate > {{.ShortLimit}}, 1,0) as short_burn_exceeded
| long_burn_exceeded + short_burn_exceeded as combined_burn
`

const TimeSliceMultiWindowMultiBurnTmpl = `_view={{.View}}
| timeslice 1m
| fillmissing timeslice(1m)
| sum(sliceGoodCount) as timesliceGoodCount, sum(sliceTotalCount) as timesliceTotalCount by _timeslice
| if(timesliceTotalCount >0, (timesliceGoodCount/timesliceTotalCount), 1)  as timesliceRatio
| if(timesliceRatio >= 0.9, 1,0) as sliceHealthy | _timeslice as _messagetime
| 1 as timesliceOne
| timeslice {{.ShortWindow}} 
| sum(sliceHealthy) as tmGood, sum(timesliceOne) as tmCount  group by _timeslice
| tmGood/tmCount as tmSLO 
| (tmCount-tmGood) as tmBad 
| total tmCount as totalCount  
| totalCount*(1-{{.Target}}) as errorBudget
| ((tmBad/tmCount)/(1-{{.Target}})) as sliceBurnRate
| if(queryEndTime() - _timeslice <= {{.ShortWindow}},sliceBurnRate, 0  )  as latestBurnRate 
| sum(tmGood) as totalGood, max(totalCount) as totalCount, max(latestBurnRate) as latestBurnRate 
| (1-(totalGood/totalCount))/(1-{{.Target}}) as longBurnRate
| if (longBurnRate > {{.LongLimit}} , 1,0) as long_burn_exceeded
| if ( latestBurnRate > {{.ShortLimit}}, 1,0) as short_burn_exceeded
| long_burn_exceeded + short_burn_exceeded as combined_burn
`

func MonitorConfigFromOpenSLO(sloConf SLO) (*SLOMonitorConfig, error) {

	var tmpl *template.Template
	var err error
	if sloConf.Spec.BudgetingMethod == BudgetingMethodNameTimeSlices {
		tmpl, err = template.New("monitor").Parse(TimeSliceMultiWindowMultiBurnTmpl)
	} else {
		tmpl, err = template.New("monitor").Parse(MultiWindowMultiBurnTmpl)
	}

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

	alertTmplParams := ConvertToBurnRateTmplParams(alerts, sloConf.Target(), sloConf.TimesliceTarget())

	for _, alert := range alertTmplParams {

		sortedNotifs := make([]Notification, len(alert.Notifications))

		copy(sortedNotifs, alert.Notifications)
		sort.Slice(sortedNotifs, func(i, j int) bool {
			return GiveStructCompare(sortedNotifs[i], sortedNotifs[j])
		})

		alert.View = sloConf.ViewName()
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

	sort.Slice(objectives, func(i, j int) bool {
		return objectives[i].Suffix < objectives[j].Suffix
	})

	mConf.Objectives = objectives
	return mConf, nil
}

type BurnAlertTmplParams struct {
	BurnRate
	Target               float64 // SLO goal i.e. the slo percent age in [0.0,1.0] decimal form
	TimesliceRatioTarget float64 // only applicable for timeslice based budgeting
}

func ConvertToBurnRateTmplParams(alerts []BurnRate, target, timesliceTarget float64) []BurnAlertTmplParams {
	var tmplAlertsParams []BurnAlertTmplParams

	for _, alert := range alerts {
		tmplAlertsParams = append(tmplAlertsParams, BurnAlertTmplParams{
			BurnRate:             alert,
			Target:               target,
			TimesliceRatioTarget: timesliceTarget,
		})
	}

	return tmplAlertsParams
}

func giveLocalTimeZone() string {
	loc, err := time.LoadLocation("Local")

	if err != nil {
		log.Fatal(err)
	}

	return loc.String()
}
