package libs

import (
	"reflect"
	"testing"
	"time"
)

func TestViewConfigFromSLO(t *testing.T) {
	s, err := Parse("templates/openslo/tsat-batcher.yaml")

	if err != nil {
		t.Errorf(err.Error())
		return
	}
	type args struct {
		sloConf SLO
	}
	tests := []struct {
		name    string
		args    args
		want    *ScheduledView
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "basic",
			args: args{
				sloConf: *s,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ViewConfigFromSLO(tt.args.sloConf)
			t.Logf("\n%+v\n\n", got.Query)

			if (err != nil) != tt.wantErr {
				t.Errorf("ViewConfigFromSLO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ViewConfigFromSLO() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStartOfMonth(t *testing.T) {
	type args struct {
		offset time.Month
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		// TODO: Add test cases.
		{
			name: "a",
			args: args{offset: -1},
			want: time.Time{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStartOfMonth(tt.args.offset); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStartOfMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}
