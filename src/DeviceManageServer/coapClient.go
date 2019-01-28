package DeviceManageServer

import (
	"encoding/json"
	"fmt"
	"github.com/dustin/go-coap"
	"math/rand"
)

const MAX_MASSAGEID = 1 << 16 //2^16

type coap_struGetTempReq struct {
	host string
}

type coap_struGetTempResp struct {
	CodeID      int `json:"CodeID"`
	Temperature int `json:"Temperature"`
}

func makeMessageID() uint16 {
	return uint16(rand.Int31() % MAX_MASSAGEID)
}

func coapclient_getTemperature(m_req *coap_struGetTempReq, m_resp *coap_struGetTempResp) error {
	req := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.GET,
		MessageID: makeMessageID(),
	}

	path := "/pet/health"

	req.SetPathString(path)

	c, err := coap.Dial("udp", m_req.host)
	if err != nil {
		logger.Error(fmt.Sprintf("Error dialing: %s", err.Error()))
		return err
	}

	rv, err := c.Send(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Error sending request: %s", err.Error()))
		return err
	}

	if rv != nil {
		logger.Info(fmt.Sprintf("Response payload: %s", rv.Payload))
		json.Unmarshal(rv.Payload, m_resp)
	}

	return err
}
