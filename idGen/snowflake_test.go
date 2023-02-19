package idGen

import (
	"testing"
)

func TestNewSnowflake(t *testing.T) {
	type args struct {
		datacenterid int64
		workerid     int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"1", args{31, 31}, false},
		{"2", args{0, 0}, false},
		{"3", args{33, 33}, false},
		{"4", args{32, 32}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSnowflake(tt.args.datacenterid, tt.args.workerid)
			if (err != nil) != tt.wantErr {
				t.Log("datacenterid max", datacenteridMax)
				t.Log("workerid max", workeridMax)
				t.Errorf("NewSnowflake() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			id, err := got.NextID()
			if (err != nil) != tt.wantErr {
				t.Errorf("got.NextID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log("id: ", id)

		})
	}
}
