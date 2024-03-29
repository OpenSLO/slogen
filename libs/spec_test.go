package libs

import (
	"gopkg.in/yaml.v3"
	"testing"
)

func TestSLO_ViewID(t *testing.T) {

	var slo SLOv1Alpha
	err := yaml.Unmarshal([]byte(testYaml1), &slo)

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	tests := []struct {
		name string
		s    SLOv1Alpha
		want string
	}{
		// TODO: Add test cases.
		{
			name: "test",
			s:    slo,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.s.ViewID(); got != tt.want {
				t.Errorf("ViewID() = %v, want %v", got, tt.want)
			}
		})
	}
}

var testYaml1 = `
apiVersion: openslo/v1alpha
kind: SLO
metadata:
  displayName: CloudCollector Ingest Lag
  name: cc-ingest-lag-v2
spec:
  service: cloudcollector
  description: Track number of seconds a message is delayed in the ingest pipeline
  budgetingMethod: Timeslices
  objectives:
    - displayName: SLI to track ingest job is completed within 5 seconds for cloudcollector
      target: 0.95
      timeSliceTarget: 0.9 # ratio of good to total msgs, so as to consider that time window healthy, only applicable for Timeslices budgeting
      ratioMetrics:
        total:
          source: sumologic
          queryType: Logs
          query: |
            _sourcecategory=cloudcollector DefaultPerCustomerLagTracker !CustomerLagQueryDisablingStrategy "current lag"
              | parse "current lag: Some(*) ms," as lag
              | where lag != "*"
              | parse "customer: *," as customer_id
              | where customer_id matches "*"
              | lag / 1000 as lag_seconds
        good:
          source: sumologic
          queryType: Logs
          query: lag_seconds <= 20
        incremental: true
createView: true
fields:
  customerID: "customer_id"
  deployment: 'if(isNull(deployment),"dev",deployment)' # using an expression
  cluster: 'if(isNull(cluster),"-",cluster)'
labels:
  team: collection
  tier: 0
alerts:
  burnRate:
    - shortWindow: '10m'
      shortLimit: 14
      longWindow: '1h'
      longLimit: 14
      notifications:
        - connectionType: 'Email'
          recipients:
            - 'agaurav@sumologic.com'
          triggerFor:
            - Warning
            - ResolvedWarning
`
