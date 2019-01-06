package DeviceManageServer

import (
	"log"
	"time"
	//	"runtime"
	"errors"
	"sync"
)

const (
	DEVNUM        = 10000 //初始化设备容量
	UPDATETIME    = 3     //该时间(s)内设备缓存数据有效
	KEEPALVIETIME = 3     //设备心跳间隔在该时间(s)内在线
)

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

type StruBusiness struct {
	devInfoMap      map[int]*struDevInfo //<codeID,*struDevInfo>
	devDataList     [DEVNUM]*struDevData //[index] *struDevData
	currDevNum      int                  //指通过app绑定的数量
	devDataReqPool  sync.Pool            //定时更新请求的对象池
	devDataRespPool sync.Pool            //定时更新回复的对象池
}

var business *StruBusiness = &StruBusiness{}

func GetBusinessInstance() *StruBusiness {
	return business
}

func (b *StruBusiness) Init() {

	b.devInfoMap = make(map[int]*struDevInfo)
	//	b.devDataList = make([]*struDevData, 0, DEVNUM)
	b.devDataReqPool = sync.Pool{
		New: func() interface{} { return new(coap_struGetTempReq) },
	}
	b.devDataRespPool = sync.Pool{
		New: func() interface{} { return new(coap_struGetTempResp) },
	}
}

func (b *StruBusiness) getDevInfo(codeID int) *struDevInfo {
	devinfo, ok := b.devInfoMap[codeID]
	if !ok {
		//Start时未加载后续绑定的设备
		if index := db_getDevIndex(codeID); index == 0 {
			log.Printf("codeID : %d is not exsit", codeID)
			return nil
		} else {
			devinfo = &struDevInfo{
				codeID: codeID,
				index:  index,
			}
			b.devInfoMap[codeID] = devinfo
			b.currDevNum++
		}
	}
	return devinfo
}

func (b *StruBusiness) getTemperature(r *struGetDevTempReq, w *struGetDevTempResp) error {

	w.CodeID = r.codeID

	if devinfo := b.getDevInfo(r.codeID); devinfo == nil {
		w.setCommonResp(DMS_ERR_DEV_NOTEXIST)
		return errors.New("device is not exsit")
	} else {
		if devinfo.status == false {
			w.setCommonResp(DMS_ERR_DEV_OFFLINE)
			return errors.New("device is offline")
		}
		if devData := b.devDataList[devinfo.index]; devData != nil && time.Since(devData.updateTime).Seconds() <= UPDATETIME {
			//有缓存且数据具有时效性
			w.Temperature = devData.temperature
		} else {
			//无命中缓存或数据源无时效性,CoAP client->
			req := b.devDataReqPool.Get().(*coap_struGetTempReq)
			defer b.devDataReqPool.Put(req)
			req.host = devinfo.host

			//用对象池代替new
			// req := &coap_struGetTempReq{
			// 	host: devinfo.host,
			// }

			resp := b.devDataRespPool.Get().(*coap_struGetTempResp)
			defer b.devDataRespPool.Put(resp)
			//resp := new(coap_struGetTempResp)
			if err := coapclient_getTemperature(req, resp); err != nil {
				w.setCommonResp(DMS_ERR_DEV_COAPFAIL)
				return errors.New("device coap fail")
			}
			w.Temperature = resp.temperature

			if devData != nil {
				//更新缓存
				devData.updateTime = time.Now()
				devData.temperature = resp.temperature
			} else {
				//加缓存
				devData = &struDevData{
					updateTime:  time.Now(),
					temperature: resp.temperature,
				}
			}
			b.devDataList[devinfo.index] = devData
		}
	}
	w.setCommonResp(DMS_ERR_SUCCESS)
	return nil
}

func (b *StruBusiness) UpdateDevData() {
	for {
		for _, devinfo := range b.devInfoMap {
			if bIsClose == true {
				return
			}
			if time.Since(devinfo.keepaliveTime).Seconds() >= KEEPALVIETIME {
				devinfo.status = false
				continue //如果设备不在线，不更新设备数据
			}

			if devData := b.devDataList[devinfo.index]; devData == nil || time.Since(devData.updateTime).Seconds() > UPDATETIME {
				req := &coap_struGetTempReq{
					host: devinfo.host,
				}
				resp := new(coap_struGetTempResp)
				if err := coapclient_getTemperature(req, resp); err != nil {
					continue
				}

				if devData != nil {
					//更新缓存
					devData.updateTime = time.Now()
					devData.temperature = resp.temperature
				} else {
					//加缓存
					devData = &struDevData{
						updateTime:  time.Now(),
						temperature: resp.temperature,
					}
				}
				b.devDataList[devinfo.index] = devData
			}
		}
	}
}

func (b *StruBusiness) Start() {
	b.currDevNum = db_getDevInfo(&(b.devInfoMap))
	go b.UpdateDevData()
}

func (b *StruBusiness) Close() {

}
