package convert

import (
	"encoding/json"
)

func StructToJSON(params interface{}) (json.RawMessage, error) {
	var jsonBs json.RawMessage
	if params != nil {
		bs, err := json.Marshal(params)
		if err != nil {
			return jsonBs, err
		}
		jsonBs = bs
	}

	return jsonBs, nil
}
