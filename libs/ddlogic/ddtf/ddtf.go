package ddtf

type SLOLibraryFolder struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     int      `json:"version"`
	CreatedAt   string   `json:"createdAt"`
	CreatedBy   string   `json:"createdBy"`
	ModifiedAt  string   `json:"modifiedAt"`
	ModifiedBy  string   `json:"modifiedBy"`
	ParentID    string   `json:"parentId"`
	ContentType string   `json:"contentType"`
	Type        string   `json:"type"`
	IsSystem    bool     `json:"isSystem"`
	IsMutable   bool     `json:"isMutable"`
	IsLocked    bool     `json:"isLocked"`
	Permissions []string `json:"permissions"`
}

type SLOLibrarySLO struct {
	ResourceName    string // terraform resource name to override the one calculated from the SLO name+service
	ID              string `json:"id,omitempty"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Version         int    `json:"version"`
	CreatedAt       string `json:"createdAt"`
	CreatedBy       string `json:"createdBy"`
	ModifiedAt      string `json:"modifiedAt"`
	ModifiedBy      string `json:"modifiedBy"`
	ParentID        string `json:"parentId"`
	MonitorFolderID string
	ContentType     string        `json:"contentType"`
	Type            string        `json:"type"`
	IsSystem        bool          `json:"isSystem"`
	IsMutable       bool          `json:"isMutable"`
	IsLocked        bool          `json:"isLocked"`
	SignalType      string        `json:"signalType"` // string^(Latency|Error|Throughput|Availability|Other)$
	Compliance      SLOCompliance `json:"compliance"`
	Indicator       SLOIndicator  `json:"indicator"`
	Service         string        `json:"service"`
	Application     string        `json:"application"`
}

type SLOCompliance struct {
	ComplianceType string  `json:"complianceType"`       // string^(Window|Request)$
	Target         float64 `json:"target"`               // [0..100]
	Timezone       string  `json:"timezone"`             // IANA Time Zone Database
	Size           string  `json:"size,omitempty"`       // Must be a multiple of days (minimum 1d, and maximum 14d)
	WindowType     string  `json:"windowType,omitempty"` // string^(Daily|Weekly|Monthly|Yearly)$
	StartFrom      string  `json:"startFrom,omitempty"`
}

type SLOIndicator struct {
	EvaluationType string          `json:"evaluationType"` // string^(Window|Request)$
	QueryType      string          `json:"queryType"`      // string^(Logs|Metrics)$
	Queries        []SLIQueryGroup `json:"queries"`
	Threshold      float64         `json:"threshold"`
	Op             string          `json:"op,omitempty"`
	Aggregation    string          `json:"aggregation,omitempty"`
	Size           string          `json:"size,omitempty"`
}

type SLI struct {
	EvaluationType string          `json:"evaluationType"` // string^(Window|Request)$
	QueryType      string          `json:"queryType"`      // string^(Logs|Metrics)$
	Queries        []SLIQueryGroup `json:"queries"`
}

type SLIQueryGroup struct {
	QueryGroupType string     `json:"queryGroupType"` // string^(Successful|Unsuccessful|Total|Threshold)$
	QueryGroup     []SLIQuery `json:"queryGroup"`
}

type SLIQuery struct {
	RowId       string `json:"rowId"`
	Query       string `json:"query"`
	UseRowCount bool   `json:"useRowCount"`
	Field       string `json:"field,omitempty"`
}

// SloBurnRateCondition struct for SloBurnRateCondition
type SloBurnRateCondition struct {
	TriggerCondition
	// The burn rate percentage.
	BurnRateThreshold float64 `json:"burnRateThreshold"`
	// The relative time range for the burn rate percentage evaluation.
	TimeRange string `json:"timeRange"`
}

type TriggerCondition struct {
	TimeRange         string  `json:"timeRange"`
	TriggerType       string  `json:"triggerType"`
	Threshold         float64 `json:"threshold,omitempty"`
	ThresholdType     string  `json:"thresholdType,omitempty"`
	OccurrenceType    string  `json:"occurrenceType"`
	TriggerSource     string  `json:"triggerSource"`
	DetectionMethod   string  `json:"detectionMethod"`
	Field             string  `json:"field,omitempty"`
	Window            int     `json:"window,omitempty"`
	BaselineWindow    string  `json:"baselineWindow,omitempty"`
	Consecutive       int     `json:"consecutive,omitempty"`
	Direction         string  `json:"direction,omitempty"`
	SLIThreshold      float64 `json:"sliThreshold,omitempty"`
	BurnRateThreshold float64 `json:"burnRateThreshold,omitempty"`
}
