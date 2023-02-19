package convert

import (
	"github.com/chenxinqun/ginWarpPkg/timex"
	"testing"
	"time"
)

type S1 struct {
	A *time.Time
	B int
	C bool
	D *T2
}
type S2 struct {
	D S1
	E *S1
	F map[string]interface{}
	G []string
}
type T1 struct {
	A *timex.JSONTime
	C bool
	D []*T2
}

type T2 struct {
	B int
	C bool
}

type T3 struct {
	D S1
	E *S1
	F map[string]interface{}
	G []string
}
type T4 struct {
	S1
}

func TestStructToStruct(t *testing.T) {
	tm := time.Now()
	jtm := timex.JSONTimeNow()
	type args struct {
		src    interface{}
		target interface{}
	}
	var tests = []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"1", args{src: &S1{A: &tm, B: 1, C: true}, target: &T1{}}, false},
		{"1", args{src: &T4{S1{A: &tm, B: 1, C: true}}, target: &S1{}}, false},
		{"1", args{src: T1{A: &jtm, C: true}, target: &S1{}}, false},
		{"1", args{src: S1{A: &tm, B: 1, C: true}, target: &T2{}}, false},
		{"1", args{src: S2{D: S1{A: &tm, B: 1, C: true}, E: &S1{A: &tm, B: 1, C: true}}, target: &T3{}}, false},
		{"1", args{src: S2{D: S1{A: &tm, B: 1, C: true}, E: &S1{A: &tm, B: 1, C: true}, F: map[string]interface{}{"a": "a"}, G: []string{"a", "b"}}, target: &T3{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := StructToStruct(tt.args.src, tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("StructToStruct() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Log(tt.args.src, tt.args.target)
			}
		})
	}
}

//
//func TestStructSliceToSlice(t *testing.T) {
//	s := make([]T3, 0)
//	ta := make([]T3, 0)
//	s = append(s, T3{})
//	err := StructSliceToSlice(s, &ta)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	log.Println(s, ta)
//}

func TestStructToQuery2(t *testing.T) {
	h, err := StructToQuery(struct {
		List []string
	}{[]string{"1", "2", "3"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(h)
}
