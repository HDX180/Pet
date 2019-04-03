package DeviceManageServer

import (
	"go.uber.org/zap"
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

type StruBusiness struct {
	devInfoMaps    [WORKNUM]sync.Map    //<codeID,*struDevInfo>
	devDataMap     map[int]*struDevData //<index, *struDevData>
	devDataMapLock sync.RWMutex
	currDevNum     int      //指通过app绑定的数量
	devAliChan     chan int //心跳队列
}

var business *StruBusiness = &StruBusiness{}

func GetBusinessInstance() *StruBusiness {
	return business
}

func (b *StruBusiness) Init() {
	b.devDataMap = make(map[int]*struDevData)
	b.devAliChan = make(chan int, 200)
}

func (b *StruBusiness) getDevDataMap(index int) (*struDevData, bool) {
	b.devDataMapLock.RLock()
	defer b.devDataMapLock.RUnlock()
	val, ok := b.devDataMap[index]
	return val, ok
}

func (b *StruBusiness) setDevDataMap(index int, v *struDevData) {
	b.devDataMapLock.Lock()
	defer b.devDataMapLock.Unlock()
	b.devDataMap[index] = v
}

func (b *StruBusiness) getDevInfoMap(codeID int) *struDevInfo {
	var index int = codeID & (WORKNUM - 1)
	val, ok := b.devInfoMaps[index].Load(codeID)
	if ok {
		return val.(*struDevInfo)
	} else {
		return nil
	}
}

func (b *StruBusiness) setDevInfoMap(codeID int, d *struDevInfo) {
	var index int = codeID & (WORKNUM - 1)
	b.devInfoMaps[index].Store(codeID, d)
}

func (b *StruBusiness) getDevInfo(codeID int) *struDevInfo {
	devinfo := b.getDevInfoMap(codeID)
	if devinfo == nil {
		//Start时未加载后续绑定的设备
		if index := db_getDevIndex(codeID); index == 0 {
			logger.Error("dev is not exsit", zap.Int("codeID", codeID))
			return nil
		} else {
			devinfo = &struDevInfo{
				codeID: codeID,
				index:  index,
			}
			devinfo.status = false
			b.setDevInfoMap(codeID, devinfo)
			b.currDevNum++
		}
	}
	return devinfo
}

func (b *StruBusiness) handleDevMsg() {
	for {
		codeID, ok := <-b.devAliChan
		if ok == false {
			break
		}
		if devinfo := b.getDevInfo(codeID); devinfo != nil {
			devinfo.keepaliveTime = time.Now()
			devinfo.status = true
		}
	}
}

func (b *StruBusiness) getPetHealth(r *struGetPetHealthReq, w *struGetPetHealthResp) {

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
		if devData, ok := b.getDevDataMap(devinfo.index); ok && time.Since(devData.updateTime).Seconds() <= UPDATETIME {
			//有缓存且数据具有时效性
			w.Temperature = devData.temperature
			intoCache++
		} else {
			//无命中缓存或数据源无时效性,CoAP client->
			logger.Info("data is out of date", zap.Int("codeID", r.codeID))
			req := &coap_struGetPetHealthReq{
				host: devinfo.host,
			}
			resp := new(coap_struGetPetHealthResp)
			if err := coapclient_getPetHealth(req, resp); err != nil {
				w.setCommonResp(DMS_ERR_DEV_COAPFAIL)
				totalReqNum--
				return
			}
			w.Temperature = resp.Temperature

			if ok {
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
			b.setDevDataMap(devinfo.index, devData)
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
		b.devInfoMaps[i].Range(func(_, val interface{}) bool {
			// for _, devinfo := range b.devInfoMaps[i] {
			if bIsClose == true {
				return false
			}

			devinfo := val.(*struDevInfo)

			if time.Since(devinfo.keepaliveTime).Seconds() >= KEEPALVIETIME {
				devinfo.status = false
				return true //如果设备不在线，不更新设备数据
			}

			if devData, ok := b.getDevDataMap(devinfo.index); !ok || time.Since(devData.updateTime).Seconds() > UPDATETIME {
				req := devDataReqPool.Get().(*coap_struGetTempReq)
				req.host = devinfo.host
				resp := devDataRespPool.Get().(*coap_struGetTempResp)

				if err := coapclient_getTemperature(req, resp); err != nil {
					devDataRespPool.Put(resp)
					devDataReqPool.Put(req)
					return true
				}

				if ok {
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
				b.setDevDataMap(devinfo.index, devData)
				devDataRespPool.Put(resp)
				devDataReqPool.Put(req)
			}
			return true
		})
		if bIsClose == true {
			return
		}
		time.Sleep(time.Duration(500) * time.Millisecond)
	}
}

func (b *StruBusiness) Start() {
	b.currDevNum, _ = db_getDevInfo()
	go b.handleDevMsg()
	for i := 0; i < WORKNUM; i++ {
		go b.UpdateDevData(i)
	}
}

func (b *StruBusiness) Close() {
	close(b.devAliChan)
}
