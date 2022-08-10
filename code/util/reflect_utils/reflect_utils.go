package reflect_utils

import (
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strings"
)

func GetStructBson(_type reflect.Type) bson.M {
	bsonValue := bson.M{}
	for i := 0; i < _type.NumField(); i++ {
		field := _type.Field(i)
		fieldBsonStringValue, ok := field.Tag.Lookup("bson")
		if ok {
			parts := strings.Split(fieldBsonStringValue, ",")
			if len(parts) > 1 {
				if parts[1] == "inline" && field.Type.Kind() == reflect.Struct {
					inlineBson := GetStructBson(field.Type)
					for k, v := range inlineBson {
						bsonValue[k] = v
					}
					continue
				}
			}
			if parts[0] != "-" && len(parts[0]) > 0 {
				bsonValue[parts[0]] = 1
			}
		}
	}
	return bsonValue
}

//根据所需的信息返回Select
func GetStructBsonCustomized(_srcType reflect.Type, _goalType reflect.Type) bson.M {
	bsonValue := bson.M{}
	_srcBsonValue := GetStructBson(_srcType)
	_goalBsonValue := GetStructBson(_goalType)
	for k := range _goalBsonValue {
		if v, ok := _srcBsonValue[k]; ok {
			bsonValue[k] = v
		}
	}
	return bsonValue
}

func IsInt(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func IsFloat(value interface{}) bool {
	switch value.(type) {
	case float32, float64:
		return true
	}
	return false
}

func ToInt(v interface{}, def int) int {
	switch v.(type) {
	case int:
		return v.(int)
	case int8:
		return int(v.(int8))
	case int16:
		return int(v.(int16))
	case int32:
		return int(v.(int32))
	case int64:
		return int(v.(int64))
	case uint:
		return int(v.(uint))
	case uint8:
		return int(v.(uint8))
	case uint16:
		return int(v.(uint16))
	case uint32:
		return int(v.(uint32))
	case uint64:
		return int(v.(uint64))
	}
	if IsFloat(v) {
		return int(v.(float64))
	}
	return def
}

func ToMapStr(v interface{}) map[string]interface{} {
	switch v.(type) {
	case map[string]interface{}:
		return v.(map[string]interface{})
	case bson.M:
		return v.(bson.M)
	}
	return nil
}

func ToArr(v interface{}) []interface{} {
	switch v.(type) {
	case []interface{}:
		return v.([]interface{})
	}
	return nil
}

func ToUint32(value interface{}, def uint32) uint32 {
	if IsInt(value) {
		return value.(uint32)
	}
	if IsFloat(value) {
		return uint32(value.(float64))
	}
	return def
}
