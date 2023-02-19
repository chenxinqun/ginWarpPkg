package convert

import (
	"testing"
)

type testQueryValues struct {
	A string `form:"a"`
	B int64  `json:"b"`
	D float64
}

func TestStructToQuery(t *testing.T) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"1", args{
			obj: testQueryValues{A: "A", B: 123456789, D: 3.1415926},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRet, err := StructToQuery(tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("StructToQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(gotRet)
		})
	}
}
