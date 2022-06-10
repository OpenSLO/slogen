package sumologic

import (
	"fmt"
	"github.com/OpenSLO/slogen/libs/specs"
	"github.com/OpenSLO/slogen/libs/sumologic/sumotf"
	"log"
	"strconv"
)

type SLOMonitor struct {
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
}

type SLO struct {
	*sumotf.SLOLibrarySLO
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
	timezone := ""
	startFrom := ""
	windowType := ""
	complianceType := "Calendar"

	if len(slo.Spec.TimeWindow) == 1 {
		if slo.Spec.TimeWindow[0].IsRolling {
			complianceType = "Rolling"
			size = slo.Spec.TimeWindow[0].Duration
		} else {
			windowType = slo.Spec.TimeWindow[0].Duration
			timezone = slo.Spec.TimeWindow[0].Calendar.TimeZone
			startFrom = slo.Spec.TimeWindow[0].Calendar.StartTime
		}
	} else {
		return nil, fmt.Errorf("no or more than one `timeWindow` for slo mentioned")
	}

	indicator, _ := giveSLI(slo)

	sumoSLO := &SLO{
		&sumotf.SLOLibrarySLO{
			Name:        slo.Spec.Indicator.Metadata.Name,
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
