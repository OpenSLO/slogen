package libs

import (
	"reflect"
	"testing"
)

func Test_giveMostCommonVars(t *testing.T) {
	type args struct {
		slos SLOMap
		n    int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
		{
			name: "a",
			args: args{
				slos: nil,
				n:    4,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := giveMostCommonVars(tt.args.slos, tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("giveMostCommonVars() = %v, want %v", got, tt.want)
			}
		})
	}
}
