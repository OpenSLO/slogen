package libs

import (
	"reflect"
	"testing"
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
