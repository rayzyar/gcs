package prebookdata

import (
	"sync"
)

const (
	StateAllocating = "Allocating"
)

type PrebookData struct {
	PreBookCode       string  `json:"preBookCode"`
	Item              string  `json:"item"`
	State             string  `json:"state"`
	WeightKG          float64 `json:"weight"`
	DriverName        string  `json:"driverName"`
	DriverPhoneNumber string  `json:"driverPhoneNumber"`
	PlateNumber       string  `json:"plateNumber"`
	PickUpTime        int64   `json:"pickUpTime"`
}

type Dao interface {
	Get(preBookingCode string) PrebookData
	StorePrebookData(pdata PrebookData)
	Delete(preBookingCode string)
	GetByToken(token string) string
}

type daoInMem struct {
	m        map[string]PrebookData
	tokenMap map[string]string
	tLock    sync.RWMutex
	mLock    sync.RWMutex
}

var DaoV1 = daoInMem{
	m:        make(map[string]PrebookData),
	tokenMap: make(map[string]string),
}

func (d *daoInMem) Get(preBookingCode string) PrebookData {
	d.mLock.RLock()
	defer d.mLock.RUnlock()
	return d.m[preBookingCode]
}

func (d *daoInMem) StorePrebookData(pdata PrebookData) {
	d.mLock.Lock()
	defer d.mLock.Unlock()
	d.m[pdata.PreBookCode] = pdata
}

func (d *daoInMem) Delete(preBookingCode string) {
	d.mLock.Lock()
	defer d.mLock.Unlock()
	delete(d.m, preBookingCode)
}

func (d *daoInMem) GetByToken(token string) string {
	d.tLock.Lock()
	defer d.tLock.Unlock()
	return d.tokenMap[token]

}
