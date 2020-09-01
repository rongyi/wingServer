package utils

import (
	"encoding/json"
	"fmt"

	"github.com/siovanus/wingServer/http/common"
)

func ParseParams(req interface{}, params map[string]interface{}) error {
	jsonData, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("ParseParams: marshal params failed, err: %s", err)
	}
	err = json.Unmarshal(jsonData, req)
	if err != nil {
		return fmt.Errorf("ParseParams: unmarshal req failed, err: %s", err)
	}
	return nil
}

func RefactorResp(resp *common.Response, errCode uint32) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return m, fmt.Errorf("RefactorResp: marhsal resp failed, err: %s", err)
	}
	err = json.Unmarshal(jsonResp, &m)
	if err != nil {
		return m, fmt.Errorf("RefactorResp: unmarhsal resp failed, err: %s", err)
	}
	m["error"] = errCode
	return m, nil
}
