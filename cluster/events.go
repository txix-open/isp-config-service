package cluster

import (
	json2 "encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"isp-config-service/entity"
	"time"
)

type ApplyLogResponse struct {
	ApplyError string
	Result     json2.RawMessage
}

func (a ApplyLogResponse) Error() string {
	return a.ApplyError
}

type DeleteModules struct {
	Ids []string
}

type ActivateConfig struct {
	ConfigId string
	Date     time.Time
}

type UpsertConfig struct {
	Config entity.Config
	Create bool
	Unsafe bool
}

type DeleteConfigs struct {
	Ids []string
}

type DeleteCommonConfig struct {
	Id string
}

type UpsertCommonConfig struct {
	Config entity.CommonConfig
	Create bool
}

type ResponseWithError struct {
	Response  interface{}
	Error     string
	ErrorCode codes.Code
}

func NewResponseErrorf(code codes.Code, format string, args ...interface{}) ResponseWithError {
	return ResponseWithError{
		Error:     fmt.Sprintf(format, args...),
		ErrorCode: code,
	}
}

func NewResponse(value interface{}) ResponseWithError {
	return ResponseWithError{
		Response: value,
	}
}
