package convert

import (
	"encoding/json"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"reflect"
	"strings"
)

func StructToMap(obj interface{}) (ret map[string]interface{}, err error) {
	defer func() {
		e := recover()
		if e != nil {
			err = errno.Errorf("%v", e)
		}
	}()
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	ret = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		// 优先用json tag
		tagName := t.Field(i).Tag.Get("json")
		// 没有就用原名
		if tagName == "" {
			tagName = t.Field(i).Name
		}
		tagNameList := strings.Split(tagName, ",")
		if len(tagNameList) > 0 {
			tagName = tagNameList[0]
		}
		ret[tagName] = v.Field(i).Interface()
	}

	return
}

func MapToStruct(obj interface{}, ret interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, ret)

	return err
}
