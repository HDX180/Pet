package DeviceManageServer

import ()

type errorCode int

const (
	DMS_ERR_SUCCESS      errorCode = iota // value --> 0
	DMS_ERR_DEV_NOTEXIST                  //设备未绑定
	DMS_ERR_DEV_OFFLINE                   //设备不在线
	DMS_ERR_DEV_COAPFAIL                  //设备通信失败
)

var errorMsg = []string{
	"success",
	"device is not exsit",
	"device is offline",
	"device coap fail",
}

type StruCommonResp struct {
	ErrNo  errorCode `json:"ErrNo"`
	ErrMsg string    `json:"ErrMsg"`
}

func (c *StruCommonResp) setCommonResp(e errorCode) {
	c.ErrNo = e
	c.ErrMsg = errorMsg[e]
}

type struGetDevTempReq struct {
	codeID int
}

type struGetDevTempResp struct {
	StruCommonResp
	CodeID      int `json:"CodeID"`
	Temperature int `json:"Temperature"`
}

// subscribe  unSubscribe公用
type struSubscribeReq struct {
	UserID int         `json:"UserID"`
	CodeID int         `json:"CodeID"`
	Topics []topicType `json:"Topics"`
}
