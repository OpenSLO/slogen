package sumologic

import (
	"fmt"
	oslo "github.com/OpenSLO/oslo/pkg/manifest/v1"
	"github.com/OpenSLO/slogen/libs/specs"
	"github.com/OpenSLO/slogen/libs/sumologic/sumotf"
	"log"
	"strconv"
	"strings"
)

const (
	MonitorKindBurnRate    = "BurnRate"
	MonitorKindSLI         = "SLI"
	ComplianceTypeCalendar = "Calendar"
	ComplianceTypeRolling  = "Rolling"
)

const (
	AnnotationEmailRecipients = "recipients"
	AnnotationEmailSubject    = "subject"
	AnnotationEmailBody       = "body"
	AnnotationEmailTimeZone   = "timezone"
	AnnotationRunForTriggers  = "run_for_triggers"
	AnnotationConnectionID    = "connection_id"
	AnnotationConnectionType  = "connection_type"
)

type SLOMonitor struct {
	Service                   string
	SLOName                   string
	SloID                     string
	MonitorName               string
	EvaluationDelay           string
	TriggerType               string
	SliThresholdWarning       float64
	SliThresholdCritical      float64
	TimeRangeWarning          string
	TimeRangeCritical         string
	BurnRateThresholdWarning  float64
	BurnRateThresholdCritical float64
	NotifyEmails              []NotifyEmail
	NotifyConnections         []NotifyConnection
}

type NotifyEmail struct {
	Recipients     []string
	Subject        string
	Body           string
	TimeZone       string
	RunForTriggers []string
}

type NotifyConnection struct {
	ID             string
	Type           string
	RunForTriggers []string
}

type SLO struct {
	*sumotf.SLOLibrarySLO
}

func (s SLO) TFResourceName() string {
	return fmt.Sprintf("sumologic_slo_%s_%s", s.Service, s.Name)
}

func (s SLOMonitor) TFResourceName() string {
	return fmt.Sprintf("sumologic_monitor_%s", s.MonitorName)
}

type SLOFolder struct {
	*sumotf.SLOLibraryFolder
}

func ConvertToSumoSLO(slo specs.OpenSLOSpec) (*SLO, error) {

	signalType := "Other"

	if slo.Metadata.Annotations["sumologic/signal-type"] != "" {
		signalType = slo.Metadata.Annotations["sumologic/signal-type"]
	}

	size := ""
	timezone := "America/New_York"
	startFrom := ""
	windowType := ""
	complianceType := ComplianceTypeCalendar

	if len(slo.Spec.TimeWindow) == 1 {
		timezone = slo.Spec.TimeWindow[0].Calendar.TimeZone
		if slo.Spec.TimeWindow[0].IsRolling {
			complianceType = ComplianceTypeRolling
			size = slo.Spec.TimeWindow[0].Duration
		} else {
			windowType = slo.Spec.TimeWindow[0].Duration
			startFrom = slo.Spec.TimeWindow[0].Calendar.StartTime
		}
	} else {
		return nil, fmt.Errorf("no or more than one `timeWindow` for slo mentioned")
	}

	indicator, _ := giveSLI(slo)

	sumoSLO := &SLO{
		&sumotf.SLOLibrarySLO{
			Name:        slo.SLO.Metadata.Name,
			Description: slo.Spec.Description,
			Service:     slo.Spec.Service,
			SignalType:  signalType,
			Compliance: sumotf.SLOCompliance{
				ComplianceType: complianceType,
				Target:         slo.Spec.Objectives[0].Target,
				Timezone:       timezone,
				Size:           size,
				WindowType:     windowType,
				StartFrom:      startFrom,
			},
			Indicator: *indicator,
		},
	}

	return sumoSLO, nil
}

func giveSLI(slo specs.OpenSLOSpec) (*sumotf.SLOIndicator, error) {
	evaluationType := ""

	switch slo.Spec.BudgetingMethod {
	case "Occurrences":
		evaluationType = "Request"
	case "Timeslices":
		evaluationType = "Window"
	default:
		log.Fatalln("budgeting method not supported", slo.Spec.BudgetingMethod)
	}

	queryType := ""
	indicator := slo.Spec.Indicator

	var queries []sumotf.SLIQueryGroup

	if indicator.Spec.RatioMetric != nil {
		switch indicator.Spec.RatioMetric.Total.MetricSource.Type {
		case "sumologic-logs":
			queryType = "Logs"
		case "sumologic-metrics":
			queryType = "Metrics"
		}

		qg := giveQueryGroup(indicator.Spec.RatioMetric.Total.MetricSource.MetricSourceSpec)
		totalQuery := &sumotf.SLIQueryGroup{
			QueryGroupType: "Total",
			QueryGroup:     []sumotf.SLIQuery{qg},
		}

		queries = append(queries, *totalQuery)

		if indicator.Spec.RatioMetric.Good != nil {
			qg := giveQueryGroup(indicator.Spec.RatioMetric.Good.MetricSource.MetricSourceSpec)
			goodQuery := &sumotf.SLIQueryGroup{
				QueryGroupType: "Successful",
				QueryGroup:     []sumotf.SLIQuery{qg},
			}
			queries = append(queries, *goodQuery)
		}

		if indicator.Spec.RatioMetric.Bad != nil {
			qg := giveQueryGroup(indicator.Spec.RatioMetric.Bad.MetricSource.MetricSourceSpec)
			badQuery := &sumotf.SLIQueryGroup{
				QueryGroupType: "Unsuccessful",
				QueryGroup:     []sumotf.SLIQuery{qg},
			}
			queries = append(queries, *badQuery)
		}
	}

	op := "LessThan"
	size := ""
	threshold := 0.0
	aggregation := ""

	if indicator.Spec.ThresholdMetric != nil {
		switch indicator.Spec.ThresholdMetric.MetricSource.Type {
		case "sumologic-logs":
			queryType = "Logs"
		case "sumologic-metrics":
			queryType = "Metrics"
		}

		specSource := indicator.Spec.ThresholdMetric.MetricSource.MetricSourceSpec
		qg := giveQueryGroup(specSource)
		query := &sumotf.SLIQueryGroup{
			QueryGroupType: "Threshold",
			QueryGroup:     []sumotf.SLIQuery{qg},
		}

		op = specSource["op"]
		aggregation = specSource["aggregation"]
		size = specSource["size"]
		if specSource["threshold"] != "" {
			var err error
			threshold, err = strconv.ParseFloat(specSource["threshold"], 64)
			if err != nil {
				log.Fatalln("threshold is not a number", specSource["threshold"])
			}
		}

		queries = append(queries, *query)
	}

	sumoIndicator := sumotf.SLOIndicator{
		EvaluationType: evaluationType,
		QueryType:      queryType,
		Queries:        queries,
		Op:             op,
		Size:           size,
		Threshold:      threshold,
		Aggregation:    aggregation,
	}

	return &sumoIndicator, nil
}

func giveQueryGroup(spec map[string]string) sumotf.SLIQuery {
	query := spec["query"]
	field := spec["field"]
	rowId := spec["row_id"]

	if rowId == "" {
		rowId = "A"
	}

	return sumotf.SLIQuery{
		RowId:       rowId,
		Query:       query,
		Field:       field,
		UseRowCount: field == "",
	}
}

func ConvertToSumoMonitor(ap oslo.AlertPolicy, slo *SLO, notifyMap map[string]oslo.AlertNotificationTarget) ([]SLOMonitor, error) {

	var mons []SLOMonitor

	notifyMails, notifyconns := giveNotifyTargets(ap, notifyMap)

	for _, c := range ap.Spec.Conditions {

		name := fmt.Sprintf("%s_%s_%s", slo.Service, slo.Name, c.Metadata.Name)

		m := SLOMonitor{
			SLOName:           slo.Name,
			Service:           slo.Service,
			MonitorName:       name,
			EvaluationDelay:   c.AlertConditionInline.Spec.Condition.AlertAfter,
			NotifyEmails:      notifyMails,
			NotifyConnections: notifyconns,
		}

		switch *c.AlertConditionInline.Spec.Condition.Kind {
		case oslo.AlertConditionTypeBurnRate:
			FillBurnRateAlert(c.AlertConditionInline.Spec, &m)
		default:
			panic(fmt.Sprintf("alert condition of this kind not supported : '%s'", c.Kind))
		}

		m.SloID = fmt.Sprintf("${sumologic_slo.%s.id}", slo.TFResourceName())

		mons = append(mons, m)
	}

	return MergeMonitors(mons), nil
}

func giveNotifyTargets(ap oslo.AlertPolicy, notifyMap map[string]oslo.AlertNotificationTarget) ([]NotifyEmail, []NotifyConnection) {
	var emailTargets []NotifyEmail
	var connTargets []NotifyConnection

	for _, n := range ap.Spec.NotificationTargets {
		target, ok := notifyMap[n.TargetRef]
		if !ok {
			log.Fatalln("notification target not found", n.TargetRef)
		}

		annotations := target.Metadata.Annotations

		if strings.ToLower(target.Spec.Target) == "email" {
			notifyMail := NotifyEmail{
				Recipients:     strings.Split(annotations[AnnotationEmailRecipients], ","),
				Subject:        annotations[AnnotationEmailSubject],
				Body:           annotations[AnnotationEmailBody],
				TimeZone:       annotations[AnnotationEmailTimeZone],
				RunForTriggers: strings.Split(annotations[AnnotationRunForTriggers], ","),
			}

			emailTargets = append(emailTargets, notifyMail)
		}

		if strings.ToLower(target.Spec.Target) == "connection" {
			notifyConn := NotifyConnection{
				Type:           annotations[AnnotationConnectionType],
				ID:             annotations[AnnotationConnectionID],
				RunForTriggers: strings.Split(annotations[AnnotationRunForTriggers], ","),
			}
			connTargets = append(connTargets, notifyConn)
		}
	}
	return emailTargets, connTargets
}

func FillBurnRateAlert(c oslo.AlertConditionSpec, m *SLOMonitor) error {

	m.TriggerType = MonitorKindBurnRate
	m.EvaluationDelay = c.Condition.AlertAfter

	switch strings.ToLower(c.Severity) {
	case "critical":
		m.BurnRateThresholdCritical = float64(c.Condition.Threshold)
		m.TimeRangeCritical = c.Condition.LookbackWindow
	case "warning":
		m.BurnRateThresholdWarning = float64(c.Condition.Threshold)
		m.TimeRangeWarning = c.Condition.LookbackWindow
	}

	return nil
}

// MergeMonitors merges multiple OpenSLO monitors critical & warning into one sumo monitor
// based on the name of the monitor.
func MergeMonitors(mons []SLOMonitor) []SLOMonitor {
	burnRateMonitors := make(map[string][]SLOMonitor)

	for _, m := range mons {

		switch m.TriggerType {
		case MonitorKindBurnRate:
			burnRateMonitors[m.MonitorName] = append(burnRateMonitors[m.MonitorName], m)
		default:
			panic(fmt.Sprintf("trigger type not supported : '%s'", m.TriggerType))
		}
	}

	mergedMonitors := mergeBurnRateMonitors(burnRateMonitors)

	return mergedMonitors
}

func mergeBurnRateMonitors(mons map[string][]SLOMonitor) []SLOMonitor {
	var mergedMonitors []SLOMonitor

	for _, m := range mons {
		if len(m) != 2 {
			panic(fmt.Sprintf("monitor %s has %d monitors, expected 2", m[0].MonitorName, len(m)))
		}

		iCrit := 0
		iWarn := 1
		if m[iCrit].BurnRateThresholdWarning != 0 {
			iCrit, iWarn = iWarn, iCrit
		}

		m[iCrit].BurnRateThresholdWarning = m[iWarn].BurnRateThresholdWarning
		m[iCrit].TimeRangeWarning = m[iWarn].TimeRangeWarning

		mergedMonitors = append(mergedMonitors, m[iCrit])
	}

	return mergedMonitors
}

func GenSLOMonitorsFromAPNames(apMap map[string]oslo.AlertPolicy, ntMap map[string]oslo.AlertNotificationTarget,
	sumoSLO SLO, slo oslo.SLO) (string, error) {

	var sloMonitors []SLOMonitor

	sloAPs := slo.Spec.AlertPolicies

	for _, apName := range sloAPs {

		ap := apMap[apName]

		mons, err := ConvertToSumoMonitor(ap, &sumoSLO, ntMap)
		if err != nil {
			return "", err
		}
		sloMonitors = append(sloMonitors, mons...)
	}

	if len(sloMonitors) == 0 {
		return "", nil
	}

	return GiveMonitorTerraform(sloMonitors)
}
