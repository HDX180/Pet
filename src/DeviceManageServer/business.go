package DeviceManageServer

import (
	"fmt"
	"sync"
	"time"
)

const (
	UPDATETIME    = 10 //该时间(s)内设备缓存数据有效
	KEEPALVIETIME = 10 //设备心跳间隔在该时间(s)内在线
	WORKNUM       = 10 //工作协程数目
)

var totalReqNum int = 0
var intoCache int = 0

type requestType int

const (
	GetTemperature requestType = iota // value --> 0
)

type struDevInfo struct {
	codeID        int
	host          string
	index         int
	version       string
	status        bool
	keepaliveTime time.Time //记录心跳间隔时间
}

type struDevData struct {
	updateTime  time.Time //记录缓存更新时间
	temperature int
}

type mapDevInfo map[int]*struDevInfo //<codeID,*struDevInfo>

type StruBusiness struct {
	devInfoMaps [WORKNUM]mapDevInfo
	devDataMap  map[int]*struDevData //<index, *struDevData>
	currDevNum  int                  //指通过app绑定的数量

}

var business *StruBusiness = &StruBusiness{}

func GetBusinessInstance() *StruBusiness {
	return business
}

func (b *StruBusiness) Init() {
	for i, _ := range b.devInfoMaps {
		b.devInfoMaps[i] = make(map[int]*struDevInfo)
	}
	b.devDataMap = make(map[int]*struDevData)
}

func (b *StruBusiness) getDevInfo(codeID int) *struDevInfo {
	devinfo, ok := b.devInfoMaps[codeID%10][codeID]
	if !ok {
		//Start时未加载后续绑定的设备
		if index := db_getDevIndex(codeID); index == 0 {
			logger.Error(fmt.Sprintf("codeID : %d is not exsit", codeID))
			return nil
		} else {
			devinfo = &struDevInfo{
				codeID: codeID,
				index:  index,
			}
			b.devInfoMaps[codeID%10][codeID] = devinfo
			b.currDevNum++
		}
	}
	return devinfo
}

func (b *StruBusiness) getTemperature(r *struGetDevTempReq, w *struGetDevTempResp) {

	w.CodeID = r.codeID

	if devinfo := b.getDevInfo(r.codeID); devinfo == nil {
		w.setCommonResp(DMS_ERR_DEV_NOTEXIST)
		return
	} else {
		if devinfo.status == false {
			w.setCommonResp(DMS_ERR_DEV_OFFLINE)
			return
		}
		totalReqNum++
		if devData := b.devDataMap[devinfo.index]; devData != nil && time.Since(devData.updateTime).Seconds() <= UPDATETIME {
			//有缓存且数据具有时效性
			w.Temperature = devData.temperature
			intoCache++
		} else {
			//无命中缓存或数据源无时效性,CoAP client->
			logger.Info(fmt.Sprintf("codeID : %d data is out of date", r.codeID))
			req := &coap_struGetTempReq{
				host: devinfo.host,
			}
			resp := new(coap_struGetTempResp)
			if err := coapclient_getTemperature(req, resp); err != nil {
				w.setCommonResp(DMS_ERR_DEV_COAPFAIL)
				totalReqNum--
				return
			}
			w.Temperature = resp.Temperature

			if devData != nil {
				//更新缓存
				devData.updateTime = time.Now()
				devData.temperature = resp.Temperature
			} else {
				//加缓存
				devData = &struDevData{
					updateTime:  time.Now(),
					temperature: resp.Temperature,
				}
			}
			b.devDataMap[devinfo.index] = devData
		}
	}
	w.setCommonResp(DMS_ERR_SUCCESS)
}

func (b *StruBusiness) UpdateDevData(i int) {
	devDataReqPool := sync.Pool{
		New: func() interface{} { return new(coap_struGetTempReq) },
	}
	devDataRespPool := sync.Pool{
		New: func() interface{} { return new(coap_struGetTempResp) },
	}
	for {
		for _, devinfo := range b.devInfoMaps[i] {
			if bIsClose == true {
				return
			}
			if time.Since(devinfo.keepaliveTime).Seconds() >= KEEPALVIETIME {
				devinfo.status = false
				continue //如果设备不在线，不更新设备数据
			}

			if devData := b.devDataMap[devinfo.index]; devData == nil || time.Since(devData.updateTime).Seconds() > UPDATETIME {
				req := devDataReqPool.Get().(*coap_struGetTempReq)
				req.host = devinfo.host
				resp := devDataRespPool.Get().(*coap_struGetTempResp)

				if err := coapclient_getTemperature(req, resp); err != nil {
					devDataRespPool.Put(resp)
					devDataReqPool.Put(req)
					continue
				}

				if devData != nil {
					//更新缓存
					devData.updateTime = time.Now()
					devData.temperature = resp.Temperature

				} else {
					//加缓存
					devData = &struDevData{
						updateTime:  time.Now(),
						temperature: resp.Temperature,
					}
				}
				b.devDataMap[devinfo.index] = devData
				devDataRespPool.Put(resp)
				devDataReqPool.Put(req)
			}
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func (b *StruBusiness) Start() {
	b.currDevNum, _ = db_getDevInfo(&b.devInfoMaps)
	for i := 0; i < 10; i++ {
		go b.UpdateDevData(i)
	}

}

func (b *StruBusiness) Close() {
	return
}
