package dataupdating

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"time"
)

type DataUpdateStatusProviderImpl struct {
	dataUpdaterImpl *DataUpdaterImpl
}

func (d DataUpdateStatusProviderImpl) GetCurrentState() State {
	return d.dataUpdaterImpl.getCurrentState()
}

func (d DataUpdateStatusProviderImpl) WaitFor(state StateType, timeout time.Duration) bool {
	return d.dataUpdaterImpl.waitFor(state, timeout)
}

func (d DataUpdateStatusProviderImpl) WaitForOKState(timeout time.Duration) bool {
	return d.dataUpdaterImpl.waitFor(OK, timeout)
}

func (d DataUpdateStatusProviderImpl) Close() error {
	d.dataUpdaterImpl.close()
	return nil
}

func NewDataUpdateStatusProviderImpl(dataUpdaterImpl *DataUpdaterImpl) DataUpdateStatusProviderImpl {
	return DataUpdateStatusProviderImpl{dataUpdaterImpl: dataUpdaterImpl}
}
