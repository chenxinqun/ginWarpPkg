package timex

import (
	"database/sql/driver"
	"time"
)

type JSONTime time.Time

var (
	cst *time.Location = time.Local
)

func init() {
	/*var err error
	if cst, err = time.LoadLocation("Asia/Shanghai"); err != nil {
		panic(err)
	}*/
}

// CSTLayout China Standard JSONTime Layout
const CSTLayout = "2006-01-02 15:04:05"

// ConvertJSONTime 类型转换
func ConvertJSONTime(t time.Time) JSONTime {
	return JSONTime(t)
}

func ConvertTime(t JSONTime) time.Time {
	return time.Time(t)
}

// JSONTimeNow 获得一个当前时间
func JSONTimeNow() JSONTime {
	return JSONTime(time.Now())
}

func (t *JSONTime) UnmarshalJSON(data []byte) (err error) {
	var now time.Time
	if string(data) != "null" || string(data) != "" {
		r := []rune(string(data))
		if string(r[0]) == `"` && string(r[len(r)-1]) == `"` {
			r = r[1 : len(r)-1]
		}
		if string(r) != "" {
			now, err = ParseCSTInLocation(string(r))
			if err != nil {
				var ns string
				ns, err = RFC3339ToCSTLayout(string(r))
				if err != nil {
					return err
				}
				now, err = ParseCSTInLocation(ns)
				if err != nil {
					return err
				}
			}
		}

	}

	*t = JSONTime(now)
	return
}

func (t JSONTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(CSTLayout)+2)
	b = append(b, []byte(`"`)...)
	// 时间有值才处理
	if time.Time(t).Unix() > 0 {
		b = time.Time(t).AppendFormat(b, CSTLayout)
	}
	b = append(b, []byte(`"`)...)
	return b, nil
}

func (t *JSONTime) Scan(value interface{}) error {
	if value == nil {
		*t = JSONTime(time.Time{})
		return nil
	}
	switch value.(type) {
	case time.Time:
		*t = JSONTime(value.(time.Time))
	case string:
		r := value.(string)
		tm, err := ParseCSTInLocation(r)
		if err != nil {
			var ns string
			ns, err = RFC3339ToCSTLayout(r)
			if err != nil {
				return err
			}
			tm, err = ParseCSTInLocation(ns)
			if err != nil {
				return err
			}
		}
		*t = JSONTime(tm)
	}

	return nil
}

func (t JSONTime) Value() (driver.Value, error) {
	if time.Time(t).Unix() == 0 {
		return nil, nil
	}
	return time.Time(t), nil
}

func (t JSONTime) String() string {
	if time.Time(t).Unix() == 0 {
		return ""
	}
	return time.Time(t).Format(CSTLayout)
}
