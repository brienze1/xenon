package pmapi

import "github.com/brienze1/xenon/internal/domain/enums"

const (
	Environment         enums.PmApi = "pm.environment"
	CollectionVariables enums.PmApi = "pm.collectionVariables"
	Globals             enums.PmApi = "pm.globals"
	Variables           enums.PmApi = "pm.variables"
	RequestUrl          enums.PmApi = "pm.request.url"
	RequestHeaders      enums.PmApi = "pm.request.headers"
	RequestBody         enums.PmApi = "pm.request.body"
	ResponseCode        enums.PmApi = "pm.response.code"
	ResponseStatus      enums.PmApi = "pm.response.status"
	ResponseHeaders     enums.PmApi = "pm.response.headers"
	ResponseJson        enums.PmApi = "pm.response.json"
	ResponseText        enums.PmApi = "pm.response.text"
	ResponseTime        enums.PmApi = "pm.response.responseTime"
	Test                enums.PmApi = "pm.test"
	Expect              enums.PmApi = "pm.expect"
	SendRequest         enums.PmApi = "pm.sendRequest"
	Require             enums.PmApi = "pm.require"
)

var MvpSubset = []enums.PmApi{
	Environment,
	CollectionVariables,
	Globals,
	Variables,
	RequestUrl,
	RequestHeaders,
	RequestBody,
	ResponseCode,
	ResponseStatus,
	ResponseHeaders,
	ResponseJson,
	ResponseText,
	ResponseTime,
	Test,
	Expect,
	SendRequest,
}
