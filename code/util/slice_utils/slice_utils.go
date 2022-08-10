package slice_utils

import "reflect"

func CreateAnyTypeSlice(slice interface{}) ([]interface{}, bool) {
	val, ok := isSlice(slice)

	if !ok {
		return nil, false
	}

	sliceLen := val.Len()

	out := make([]interface{}, sliceLen)

	for i := 0; i < sliceLen; i++ {
		out[i] = val.Index(i).Interface()
	}

	return out, true
}

func isSlice(arg interface{}) (val reflect.Value, ok bool) {
	val = reflect.ValueOf(arg)

	if val.Kind() == reflect.Slice {
		ok = true
	}

	return
}

func IsInSliceInt(s []int, item int) bool {
	for i:=0;i<len(s);i++{
		if s[i] == item {
			return true
		}
	}
	return false
}

func IsInSliceInt64(s []int64, item int64) bool {
	for i:=0;i<len(s);i++{
		if s[i] == item {
			return true
		}
	}
	return false
}

func StringIndexOf(list []string, target string) int {
	index := -1
	for i, v := range list {
		if v == target {
			index = i
			break
		}
	}
	return index
}


func FilterSliceStable(x interface{}, filter func(i int) bool)  {
	val := reflect.ValueOf(x)
	if val.Kind() != reflect.Ptr {
		return
	}
	val = val.Elem()
	if val.Kind() != reflect.Slice {
		return
	}
	swap := reflect.Swapper(val.Interface())
	l, r, length := 0, 0, val.Len()
	for r < length {
		ok := filter(r)
		if !ok {
			swap(l, r)
			l++
		}
		r++
	}
	val.SetLen(l)
}