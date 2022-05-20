package orm

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

//Insert in can be *User, []*User, map[string]interface{}
func (q *Query) Insert(in interface{}) (int64, error) {
	var keys, values []string
	v := reflect.ValueOf(in)
	//剥离指针
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		keys, values = sKV(v)
	case reflect.Map:
		keys, values = mKV(v)
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			//Kind是切片时，可以用Index()方法遍历
			sv := v.Index(i)
			for sv.Kind() == reflect.Ptr || sv.Kind() == reflect.Interface {
				sv = sv.Elem()
			}
			//切片元素不是struct或者指针，报错
			if sv.Kind() != reflect.Struct {
				return 0, errors.New("method Insert error: in slice is not structs")
			}
			//keys只保存一次就行，因为后面的都一样了
			if len(keys) == 0 {
				keys, values = sKV(sv)
				continue
			}
			_, val := sKV(sv)
			values = append(values, val...)
		}
	default:
		return 0, errors.New("method Insert error: type error")
	}
	//already done
	kl := len(keys)
	vl := len(values)
	if kl == 0 || vl == 0 {
		return 0, errors.New("method Insert error: no data")
	}
	var insertValue string
	//插入多条记录时需要用","拼接一下values
	if kl < vl {
		var tmpValues []string
		for kl <= vl {
			if kl%(len(keys)) == 0 {
				tmpValues = append(tmpValues, fmt.Sprintf("(%s)", strings.Join(values[kl-len(keys):kl], ",")))
			}
			kl++
		}
		insertValue = strings.Join(tmpValues, ",")
	} else {
		insertValue = fmt.Sprintf("(%s)", strings.Join(values, ","))
	}
	query := fmt.Sprintf(`insert into %s (%s) values %s`, q.table, strings.Join(keys, ","), insertValue)
	log.Printf("insert sql: %s", query)
	st, err := q.db.Prepare(query)
	if err != nil {
		return 0, err
	}
	result, err := st.Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
