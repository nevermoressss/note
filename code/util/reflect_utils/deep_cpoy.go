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
