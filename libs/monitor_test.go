package libs

import (
	"reflect"
	"testing"
)

func TestMonitorConfigFromOpenSLO(t *testing.T) {
	type args struct {
		sloConf SLO
	}
	tests := []struct {
		name    string
		args    args
		want    []SLOMonitorConfig
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MonitorConfigFromOpenSLO(tt.args.sloConf)
			if (err != nil) != tt.wantErr {
				t.Errorf("MonitorConfigFromOpenSLO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MonitorConfigFromOpenSLO() got = %v, want %v", got, tt.want)
			}
		})
	}
}
