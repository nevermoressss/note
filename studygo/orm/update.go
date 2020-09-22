package orm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

//Update src can be *user, user, map[string]interface{}, string
func (q *Query) Update(src interface{}) (int64, error) {
	if len(q.errs) != 0 {
		return 0, errors.New(strings.Join(q.errs, ""))
	}
	v := reflect.ValueOf(src)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	var toBeUpdated, where string
	var keys, values []string
	switch v.Kind() {
	case reflect.String:
		toBeUpdated = src.(string)
	case reflect.Struct:
		keys, values = sKV(v)
	case reflect.Map:
		keys, values = mKV(v)
	default:
		return 0, errors.New("method Update error: type error")
	}
	if toBeUpdated == "" {
		if len(keys) != len(values) {
			return 0, errors.New("method Update error: keys not match values")
		}
		var kvs []string
		for idx, key := range keys {
			kvs = append(kvs, fmt.Sprintf("%s = %s", key, values[idx]))
		}
		toBeUpdated = strings.Join(kvs, ",")
	}
	if len(q.wheres) > 0 {
		where = fmt.Sprintf(`where %s`, strings.Join(q.wheres, " and "))
	}
	query := fmt.Sprintf("update %s set %s %s", q.table, toBeUpdated, where)
	st, err := q.db.Prepare(query)
	if err != nil {
		return 0, err
	}
	result, err := st.Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
