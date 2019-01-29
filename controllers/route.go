package controllers

import (
	"encoding/json"

	"isp-config-service/socket"
)

func GetRoutes() (map[string]interface{}, error) {
	bytes, err := json.Marshal(socket.GetRoutes().Routes)
	if err != nil {
		return nil, createUnknownError(err)
	}
	var response map[string]interface{}
	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return nil, createUnknownError(err)
	}
	return response, nil
}
