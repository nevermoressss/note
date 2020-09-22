package orm

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"strings"
	"time"
)

//Query will build a sql
type Query struct {
	db     *sql.DB
	table  string
	wheres []string
	only   []string
	limit  string
	offset string
	order  string
	errs   []string
}


func Connect(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	//设置连接池
	conn.SetMaxOpenConns(100)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(10 * time.Minute)
	return conn, conn.Ping()
}

func Table(db *sql.DB, tableName string) func() *Query {
	return func() *Query {
		return &Query{
			db:    db,
			table: tableName,
		}
	}
}



func sKV(v reflect.Value) ([]string, []string) {
	var keys, values []string
	t := v.Type()
	for n := 0; n < t.NumField(); n++ {
		tf := t.Field(n)
		vf := v.Field(n)
		//忽略非导出字段
		if tf.Anonymous {
			continue
		}
		//忽略无效、零值字段
		if !vf.IsValid() || reflect.DeepEqual(vf.Interface(), reflect.Zero(vf.Type()).Interface()) {
			continue
		}
		for vf.Type().Kind() == reflect.Ptr {
			vf = vf.Elem()
		}
		//有时候根据需求会组合struct，这里处理下，支持获取嵌套的struct tag和value
		//如果字段值是time类型之外的struct，递归获取keys和values
		if vf.Kind() == reflect.Struct && tf.Type.Name() != "Time" {
			cKeys, cValues := sKV(vf)
			keys = append(keys, cKeys...)
			values = append(values, cValues...)
			continue
		}
		//根据字段的json tag获取key，忽略无tag字段
		key := strings.Split(tf.Tag.Get("json"), ",")[0]
		if key == "" {
			continue
		}
		value := format(vf)
		if value != "" {
			keys = append(keys, key)
			values = append(values, value)
		}
	}
	return keys, values
}

func format(v reflect.Value) string {
	//断言出time类型直接转unix时间戳
	if t, ok := v.Interface().(time.Time); ok {
		return fmt.Sprintf("FROM_UNIXTIME(%d)", t.Unix())
	}
	switch v.Kind() {
	case reflect.String:
		return fmt.Sprintf(`'%s'`, v.Interface())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return fmt.Sprintf(`%d`, v.Interface())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf(`%f`, v.Interface())
	//如果是切片类型，遍历元素，递归格式化成"(, , , )"形式
	case reflect.Slice:
		var values []string
		for i := 0; i < v.Len(); i++ {
			values = append(values, format(v.Index(i)))
		}
		return fmt.Sprintf(`(%s)`, strings.Join(values, ","))
	//接口类型剥一层递归
	case reflect.Interface:
		return format(v.Elem())
	}
	return ""
}

func mKV(v reflect.Value) ([]string, []string) {
	var keys, values []string
	//获取map的key组成的切片
	mapKeys := v.MapKeys()
	for _, key := range mapKeys {
		value := format(v.MapIndex(key))
		if value != "" {
			values = append(values, value)
			keys = append(keys, key.Interface().(string))
		}
	}
	return keys, values
}
