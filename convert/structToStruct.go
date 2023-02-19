package convert

import (
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"github.com/chenxinqun/ginWarpPkg/timex"
	"os"
	"reflect"
	"time"
)

// StructToStruct 传入两个结构体, src可以传指针也可以传值, target必须传指针. 使用要求必须同字段,且要同类型.
func StructToStruct(src interface{}, target interface{}) (err error) {
	/*
		只支持结构体的处理
		没法处理嵌套结构体
		使用用来转换单层结构体的值.
		多层结构体还是需要手动赋值.
		字段类型为 prt, slice, array, map, interface 和 匿名struct的字段无法处理.
		如:
		type A struct {
		A []int
		B [5]int
		C map[string]interface
		D interface{}
		E struct {
				Name string
			}
		F *B
		}
		type B Struct {
			Name string
		}
		这些字段类型都无法处理, 一般都需要手动赋值, 除非两个复杂类型完全一致.
	*/
	defer func() {
		e := recover()
		if e != nil {
			err = errno.Errorf("%v", e)
		}
	}()
	sTypeOf := reflect.TypeOf(src)
	sValueOf := reflect.ValueOf(src)
	if sTypeOf.Kind() == reflect.Ptr {
		sValueOf = sValueOf.Elem()
	}

	tTypeOf := reflect.TypeOf(target).Elem()
	tValueOf := reflect.ValueOf(target).Elem()
	num := tTypeOf.NumField()
	for i := 0; i < num; i++ {
		field := tTypeOf.Field(i)
		tValue := tValueOf.Field(i)
		name := field.Name
		typeName := field.Type.Name()
		sValue := sValueOf.FieldByName(name)
		k := sValue.Kind()
		if k != reflect.Invalid {
			sTypeName := sValue.Type().Name()
			vValid := sValue.IsValid()
			tValid := typeName != "" && sTypeName != ""
			_, tTime := tValue.Interface().(time.Time)
			_, tJTime := tValue.Interface().(timex.JSONTime)
			_, tTimePrt := tValue.Interface().(*time.Time)
			_, tJTimePrt := tValue.Interface().(*timex.JSONTime)
			_, sTime := sValue.Interface().(time.Time)
			_, sJTime := sValue.Interface().(timex.JSONTime)
			_, sTimePrt := sValue.Interface().(*time.Time)
			_, sJTimePrt := sValue.Interface().(*timex.JSONTime)
			if vValid && tValid && sTypeName == typeName {
				tValueOf.Field(i).Set(sValue)
				// 复杂类型也会尝试去处理, 但是必须要类型完全一致才能处理成功.
			} else if sJTime && tTime { // timex.JSONTime => time.time
				tValueOf.Field(i).Set(reflect.ValueOf(timex.ConvertTime(sValue.Interface().(timex.JSONTime))))
			} else if sJTime && tTimePrt { // timex.JSONTime => *time.time
				tm := timex.ConvertTime(sValue.Interface().(timex.JSONTime))
				tValueOf.Field(i).Set(reflect.ValueOf(&tm))
			} else if sJTime && tTimePrt { // timex.JSONTime => *timex.JSONTime
				tm := timex.ConvertTime(sValue.Interface().(timex.JSONTime))
				tValueOf.Field(i).Set(reflect.ValueOf(&tm))
			} else if sJTimePrt && tTime { // *timex.JSONTime => time.Time
				tjm := sValue.Interface().(*timex.JSONTime)
				if tjm != nil {
					tm := timex.ConvertTime(*tjm)
					tValueOf.Field(i).Set(reflect.ValueOf(tm))
				}
			} else if sJTimePrt && tTime { // *timex.JSONTime => timex.JSONTime
				tjm := sValue.Interface().(*timex.JSONTime)
				if tjm != nil {
					tValueOf.Field(i).Set(reflect.ValueOf(*tjm))
				}

			} else if sJTimePrt && tTime { // *timex.JSONTime => *time.Time
				tjm := sValue.Interface().(*timex.JSONTime)
				if tjm != nil {
					tm := timex.ConvertTime(*tjm)
					tValueOf.Field(i).Set(reflect.ValueOf(&tm))
				}
			} else if sTime && tJTime { // time.Time => timex.JSONTime
				tValueOf.Field(i).Set(reflect.ValueOf(timex.ConvertJSONTime(sValue.Interface().(time.Time))))
			} else if sTime && tTimePrt { // time.Time => *time.Time
				tm := sValue.Interface().(time.Time)
				tValueOf.Field(i).Set(reflect.ValueOf(&tm))
			} else if sTime && tTimePrt { // time.Time => *timex.JSONTime
				tm := sValue.Interface().(time.Time)
				tjm := timex.ConvertJSONTime(tm)
				tValueOf.Field(i).Set(reflect.ValueOf(&tjm))
			} else if sTimePrt && tJTime { // *time.Time => timex.JSONTime
				tmp := sValue.Interface().(*time.Time)
				if tmp != nil {
					tm := timex.ConvertJSONTime(*tmp)
					tValueOf.Field(i).Set(reflect.ValueOf(tm))
				}

			} else if sTimePrt && tJTimePrt { // *time.Time => *timex.JSONTime
				tm := sValue.Interface().(*time.Time)
				if tm != nil {
					tjm := timex.ConvertJSONTime(*tm)
					tValueOf.Field(i).Set(reflect.ValueOf(&tjm))
				}

			} else if sTimePrt && tTime { // *time.Time => time.Time
				tm := sValue.Interface().(*time.Time)
				if tm != nil {
					tValueOf.Field(i).Set(reflect.ValueOf(*tm))

				}
			} else if !tValid && sTypeName == typeName {
				func() {
					defer func() {
						e := recover()
						if e != nil {
							_, _ = os.Stderr.Write([]byte(fmt.Sprintf("有字段传值失败, 请注意手动赋值: %v\n", e)))
						}
					}()
					tValueOf.Field(i).Set(sValue)
				}()
			}
		}
	}

	return err
}

//
//func StructSliceToSlice(src interface{}, target interface{}) (err error) {
//	defer func() {
//		e := recover()
//		if e != nil {
//			a, ok := e.(string)
//			if ok {
//				err = errors.New(a)
//			} else {
//				err, _ = e.(error)
//			}
//		}
//	}()
//
//	sTypeOf := reflect.TypeOf(src)
//	sValueOf := reflect.ValueOf(src)
//	if sTypeOf.Kind() == reflect.Ptr {
//		sValueOf = sValueOf.Elem()
//	}
//
//	tTypeOf := reflect.TypeOf(target)
//	switch tTypeOf.Kind() {
//	case reflect.Ptr:
//		tTypeOf = tTypeOf.Elem()
//	}
//	tValueOf := reflect.ValueOf(target)
//	tPtr := false
//	switch tTypeOf.Kind() {
//	case reflect.Ptr:
//		tTypeOf = tTypeOf.Elem()
//	}
//	switch tTypeOf.Kind() {
//	case reflect.Slice:
//		switch tValueOf.Kind() {
//		case reflect.Ptr:
//			tPtr = true
//
//		}
//	}
//	if !(sTypeOf.Kind() == reflect.Slice && tTypeOf.Kind() == reflect.Slice) {
//		return errors.New("src 与 target 都必须为Slice类型")
//	}
//	if sValueOf.Len() == 0 {
//		return errors.New("src 长度不能为0")
//	}
//
//	valueSlice := tValueOf
//	if tPtr {
//		valueSlice = tValueOf.Elem()
//	}
//	for i := 0; i < sValueOf.Len(); i++ {
//		vv := sValueOf.Index(i)
//		if vv.Kind() != reflect.Struct {
//			return errors.New("src 与 target 的成员都必须为Struct类型")
//		}
//		aa := reflect.New(tTypeOf.Elem())
//		err = StructToStruct(vv.Interface(), aa.Interface())
//
//		valueSlice = reflect.Append(valueSlice, aa.Elem())
//	}
//	fmt.Println(1111, tValueOf, tPtr, valueSlice.UnsafeAddr())
//	tValueOf.Set(valueSlice)
//	fmt.Println(target)
//	return nil
//}
