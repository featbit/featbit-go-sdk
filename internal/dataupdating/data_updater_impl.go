package dataupdating

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/util/log"
	"sync"
	"time"
)

const defaultListenerNums = 10

type DataUpdaterImpl struct {
	storage      DataStorage
	currentState State
	lock         sync.Mutex
	listeners    []chan State
}

func NewDataUpdaterImpl(storage DataStorage) *DataUpdaterImpl {
	return &DataUpdaterImpl{storage: storage,
		currentState: INITIALIZINGState(),
	}
}

func (d *DataUpdaterImpl) handleErrorFromStorage(errorType string, err error) {
	log.LogError("FB GO SDK: Data Storage error: {}, DataSynchronizer will attempt to receive the data", err.Error())
	d.UpdateStatus(INTERRUPTEDState(errorType, err.Error()))
}

func (d *DataUpdaterImpl) Init(allDate map[Category]map[string]Item, version int64) bool {
	if err := d.storage.Init(allDate, version); err != nil {
		d.handleErrorFromStorage(DataStorageInitError, err)
		return false
	}
	return true
}

func (d *DataUpdaterImpl) Upsert(category Category, key string, item Item, version int64) bool {
	var ret bool
	var err error
	if ret, err = d.storage.Upsert(category, key, item, version); err != nil {
		d.handleErrorFromStorage(DataStorageUpdateError, err)
		return false
	}
	return ret
}

func (d *DataUpdaterImpl) StorageInitialized() bool {
	return d.storage.IsInitialized()
}

func (d *DataUpdaterImpl) GetVersion() int64 {
	return d.storage.GetVersion()
}

func (d *DataUpdaterImpl) UpdateStatus(state State) {
	if state.StateType == "" {
		return
	}
	d.lock.Lock()
	lastState := d.currentState
	lastStateSince := d.currentState.StateSince
	lastError := d.currentState.ErrorTrack
	newStateType := state.StateType
	if newStateType == INTERRUPTED && lastState.StateType == INITIALIZING {
		newStateType = INITIALIZING
	}
	if newStateType != lastState.StateType {
		lastStateSince = time.Now()
	}
	if state.ErrorTrack.ErrorType != "" {
		lastError = state.ErrorTrack
	}
	d.currentState = State{
		StateType:  newStateType,
		StateSince: lastStateSince,
		ErrorTrack: lastError,
	}

	var chs []chan State
	if len(d.listeners) > 0 {
		chs = make([]chan State, len(d.listeners))
		copy(chs, d.listeners)
	}
	d.lock.Unlock()
	// broadcast current state
	for _, ch := range chs {
		ch <- d.currentState
	}
}

func (d *DataUpdaterImpl) getCurrentState() State {
	return d.currentState
}

func (d *DataUpdaterImpl) waitFor(state StateType, timeout time.Duration) bool {
	d.lock.Lock()
	if d.currentState.StateType == state {
		d.lock.Unlock()
		return true
	}
	if d.currentState.StateType == OFF {
		d.lock.Unlock()
		return false
	}
	// register listener
	listener := make(chan State, defaultListenerNums)
	d.listeners = append(d.listeners, listener)
	// defer remove listener
	defer func(ch chan State) {
		d.lock.Lock()
		defer d.lock.Unlock()
		chs := d.listeners
		for i, ch := range chs {
			if ch == listener {
				copy(chs[i:], chs[i+1:])
				chs[len(chs)-1] = nil
				d.listeners = chs[:len(chs)-1]
				close(ch)
				break
			}
		}
	}(listener)
	d.lock.Unlock()
	// timeout process
	var deadline <-chan time.Time
	if timeout > 0 {
		deadline = time.After(timeout)
	}
	for {
		select {
		case newState, ok := <-listener:
			if !ok {
				return false
			}
			if newState.StateType == state {
				return true
			}
			if newState.StateType == OFF {
				return false
			}
		case <-deadline:
			return false
		}
	}
}

func (d *DataUpdaterImpl) close() {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, listener := range d.listeners {
		close(listener)
	}
	d.listeners = nil
}
