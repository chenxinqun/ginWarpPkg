package convert

import (
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/errno"
	httpURL "net/url"
	"reflect"
	"strings"
)

// StructToQuery 支持 结构体和map[string]interface{}
func StructToQuery(obj interface{}) (ret httpURL.Values, err error) {
	defer func() {
		e := recover()
		if e != nil {
			err = errno.Errorf("%v", e)
		}
	}()
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	ret = make(httpURL.Values)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errno.NewError("只支持传结构体进来")
	}
	for i := 0; i < t.NumField(); i++ {
		// 优先看有没有form标签
		tagName := t.Field(i).Tag.Get("form")
		// 没有form就用json
		if tagName == "" {
			tagName = t.Field(i).Tag.Get("json")
		}
		// 都没有就用原名
		if tagName == "" {
			tagName = t.Field(i).Name
		}
		tagNameList := strings.Split(tagName, ",")
		if len(tagNameList) > 0 {
			tagName = tagNameList[0]
		}
		var vsList = make([]string, 0)
		vf := v.Field(i)
		if vf.Kind() == reflect.Slice || vf.Kind() == reflect.Array {
			for j := 0; j < vf.Len(); j++ {
				f := vf.Index(j)
				vs := fmt.Sprintf("%v", f.Interface())
				vsList = append(vsList, vs)
			}

		} else {
			vs := fmt.Sprintf("%v", v.Field(i).Interface())
			vsList = append(vsList, vs)
		}

		kv, ok := ret[tagName]
		if ok {
			ret[tagName] = append(kv, vsList...)
		} else {
			ret[tagName] = vsList
		}
	}

	return ret, err
}
