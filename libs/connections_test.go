package libs

import (
	"reflect"
	"testing"
)

func Test_GiveConnectionIDS(t *testing.T) {
	tests := []struct {
		name    string
		want    []MonitorConnections
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "basic",
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GiveConnectionIDS("")
			if (err != nil) != tt.wantErr {
				t.Errorf("giveConnectionIDS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("giveConnectionIDS() got = %v, want %v", got, tt.want)
			}
		})
	}
}
