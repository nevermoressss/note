/*
@Time : 2020/11/13 下午2:28
@Author : kevin.dww
@File : deep_cpoy.go
@Description: 基于json序列化和反序列化实现的对象深拷贝
*/
package reflect_utils

import (
	"github.com/json-iterator/go"
)

func DeepCopy(toValue interface{}, fromValue interface{}) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	buf, err := json.Marshal(fromValue)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, toValue)
}
