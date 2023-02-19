package httpDiscover

import (
	"encoding/json"
	mux "github.com/chenxinqun/ginWarpPkg/httpx/mux"
	"testing"
)

func TestCreateSpecialHttpClient(t *testing.T) {
	type args struct {
		rs []Resource
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "aaa", args: args{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateSpecialHttpClient(tt.args.rs...)
			t.Log(got)
			ctx := mux2.CreateSpecialContext(mux2.SpecialContextResource{
				UserID:   -1,
				UserName: "-1",
				TenantID: -1,
				IsAdmin:  false,
				RoleType: -1,
			})
			ret := make(map[string]interface{})
			code, err := got.GetJson(ctx, "fc-intrusion", "/system/health", struct{}{}, &ret)
			if err != nil {
				t.Fatal(err)
			}
			r, _ := json.Marshal(ret)
			t.Log(code, r)

		})
	}
}
