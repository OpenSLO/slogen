package libs

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"strings"
	"time"
)
import "text/template"

// todo create a lookup for name to display name and other metadata

const DefaultScheduledViewRetention = 100

var useViewID = false

func (s SLO) ViewName() string {
	sloName := s.Metadata.Name
	srvName := s.Spec.Service
	srvName = strings.Replace(srvName, "-", "_", -1)
	sloName = strings.Replace(sloName, "-", "_", -1)

	viewName := fmt.Sprintf("%s_%s_%s", ViewPrefix, srvName, sloName)

	if useViewID {
		return fmt.Sprintf("%s_%s", viewName, s.ViewID())
	}

	return viewName
}

// ViewID return the short sha1 hash for view query to ensure old view data is not reused
func (s SLO) ViewID() string {
	viewQuery, err := s.ViewQuery()

	if err != nil {
		log.Fatal("error creating view query", err)
	}

	viewHash := sha1.Sum([]byte(viewQuery))
	return fmt.Sprintf("%x", viewHash)[:8]
}

// ScheduledView implicit
type ScheduledView struct {
	SLOName        string `yaml:"sloName"`
	Service        string `yaml:"service"`
	Index          string `yaml:"index"`
	Query          string `yaml:"query"`
	StartTime      string `yaml:"startTime"` /// e.g. "2019-09-01T00:00:00Z"
	Retention      int    `yaml:"retention"` // unit: days
	PreventDestroy bool   `yaml:"preventDestroy,omitempty"`
}

type ViewTemplateValues struct {
	Name       string
	Service    string
	TotalQuery string
	GoodQuery  string
	Labels     map[string]string `yaml:"labels,omitempty"`
	Fields     map[string]string `yaml:"fields,omitempty"`
	Goal       float64
}

type ViewSpec struct {
	IndexName       string `yaml:"indexName"`
	AutoParse       bool   `yaml:"autoParse"`
	StartTime       string `yaml:"startTime"`
	RetentionInDays int    `yaml:"retentionInDays"`
	PreventDestroy  bool   `yaml:"preventDestroy"`
}

const ScheduledViewQueryTemp = `{{.TotalQuery}} 
| timeslice 1m
| if ( {{.GoodQuery}}, 1, 0) as isGood
{{ range $key, $val := .Fields }}
| {{ $val }} as {{ $key }}
{{ end }}
| sum(isGood) as sliceGoodCount, count as sliceTotalCount 
  by _timeslice{{ range $key, $val := .Fields }}, {{$key}}{{end}}

| "{{.Name}}" as SLOName 
| "{{.Service}}" as Service
{{ range $key, $val := .Labels }}
| "{{ $val }}" as {{ $key }}
{{- end }}
`

func (s SLO) ViewQuery() (string, error) {
	sloName := s.Metadata.Name
	tmpl, err := template.New("view-" + sloName).Parse(ScheduledViewQueryTemp)
	if err != nil {
		return "", err
	}
	goal := *(s.Spec.Objectives[0].BudgetTarget) * 100.0
	buf := bytes.Buffer{}
	ratio := s.Spec.Objectives[0].RatioMetrics

	viewVals := ViewTemplateValues{
		Name:       sloName,
		Service:    s.Spec.Service,
		TotalQuery: ratio.Total.Query,
		GoodQuery:  ratio.Good.Query,
		Fields:     s.Fields,
		Labels:     s.Labels,
		Goal:       goal,
	}

	err = tmpl.Execute(&buf, viewVals)
	if err != nil {
		return "", err
	}

	return buf.String(), nil

}

func ViewConfigFromSLO(sloConf SLO) (*ScheduledView, error) {

	viewQuery, err := sloConf.ViewQuery()
	if err != nil {
		return nil, err
	}

	//start := GetStartOfMonth().Add(-1 * time.Hour * 24 * 30).Format(time.RFC3339)
	start := GetStartOfMonth()

	// if less than 15 days from start of month then subtract 15 more days
	if time.Since(start) < 15*24*time.Hour {
		start = start.Add(-15 * 24 * time.Hour)
	}

	conf := &ScheduledView{
		SLOName:        sloConf.Name(),
		Service:        sloConf.Spec.Service,
		Index:          sloConf.ViewName(),
		Query:          viewQuery,
		StartTime:      start.UTC().Format(time.RFC3339),
		Retention:      DefaultScheduledViewRetention,
		PreventDestroy: false,
	}

	return conf, nil
}

func GetStartOfMonth() time.Time {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	return firstOfMonth
}

func GiveScheduleViewName(s SLO) string {

	sloName := s.Metadata.Name
	srvName := s.Spec.Service
	sloName = strings.Replace(sloName, "-", "_", -1)
	srvName = strings.Replace(srvName, "-", "_", -1)

	return fmt.Sprintf("%s_%s_%s", ViewPrefix, srvName, sloName)
}
