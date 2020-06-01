package cluster

import (
	json2 "encoding/json"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"isp-config-service/entity"
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

type BroadcastEvent struct {
	Event        string
	ModuleNames  []string
	PerformUntil time.Time //stateless events, must be broadcast until this in UTC
	Payload      json2.RawMessage
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
