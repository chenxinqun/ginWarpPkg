package timex

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestRFC3339ToCSTLayout(t *testing.T) {
	t.Log(RFC3339ToCSTLayout("2020-11-08T08:18:46+08:00"))
}

func TestCSTLayoutString(t *testing.T) {
	t.Log(CSTLayoutString())
}

func TestCSTLayoutStringToUnix(t *testing.T) {
	t.Log(CSTLayoutStringToUnix("2020-01-24 21:11:11"))
}

func TestGMTLayoutString(t *testing.T) {
	t.Log(GMTLayoutString())
}

func TestJSONTime_MarshalJSON(t *testing.T) {
	var jt = JSONTime{Time: time.Now()}
	a, err := json.Marshal(struct {
		Time JSONTime
	}{
		Time: jt,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(a), err)
}

func TestJSONTime_UnmarshalJSON(t *testing.T) {
	s := `{"Time":"2022-03-09 18:48:29"}`
	var jt = struct{ Time JSONTime }{}
	err := json.Unmarshal([]byte(s), &jt)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(jt)
	t.Log(jt)
}

func TestConvertJSONTime(t *testing.T) {
	t.Log(ConvertJSONTime(time.Now()))
}

func TestJSONTimeNow(t *testing.T) {
	t.Log(JSONTimeNow())
}
