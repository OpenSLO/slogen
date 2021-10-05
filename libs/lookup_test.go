package libs

import "testing"

func TestUploadSLOLookup(t *testing.T) {
	type args struct {
		id       string
		url      string
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "check",
			args: args{
				//id:  "000000000",
				url: "https://api.sumologic.com/api/v1/lookupTables/000000000/upload",
				//url:      "https://api.sumologic.com/api/v1/lookupTables/000000000",
				filename: "templates/slogen_lookup.csv",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UploadSLOLookup(tt.args.id, tt.args.url, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("UploadSLOLookup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
