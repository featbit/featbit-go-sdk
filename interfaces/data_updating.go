package interfaces

import (
	"fmt"
	"io"
	"time"
)

const (
	DataStorageInitError   = "Data Storage init error"
	DataStorageUpdateError = "Data Storage update error"
	RequestInvalidError    = "Request invalid"
	DataInvalidError       = "Received Data invalid"
	WebsocketError         = "WebSocket error"
	WebsocketCloseTimeout  = "WebSocket close timeout"
	UnknownError           = "Unknown error"
	NetworkError           = "Network error"
	UnknownCloseCode       = "Unknown close code"
)

type StateType string

const (
	// INITIALIZING the initial state of the data update processing
	// when the SDK is being initialized.
	// If it encounters an error that requires it to retry initialization, the state will remain at
	// INITIALIZING until it either succeeds and becomes OK, or permanently fails and becomes OFF.
	INITIALIZING StateType = "INITIALIZING"

	// OK indicates that the update processing is currently operational and has not had any problems since the
	// last time it received data.
	// In streaming mode, this means that there is currently an open stream connection and that at least
	// one initial message has been received on the stream.
	OK StateType = "OK"

	// INTERRUPTED indicates that the update processing encountered an error that it will attempt to recover from.
	// In streaming mode, this means that the stream connection failed, or had to be dropped due to some
	// other error, and will be retried after a backoff delay.
	INTERRUPTED StateType = "INTERRUPTED"

	// OFF indicates that the update processing has been permanently shut down.
	// This could be because it encountered an unrecoverable error or because the SDK client was
	// explicitly shut down.
	OFF StateType = "OFF"
)

type ErrorTrack struct {
	ErrorType string
	Message   string
}

func (et ErrorTrack) String() string {
	return fmt.Sprintf(`{"errorType": "%s", "message": "%s"}`, et.ErrorType, et.Message)
}

type State struct {
	StateType  StateType
	StateSince time.Time
	ErrorTrack ErrorTrack
}

func (s State) String() string {
	timeStr := s.StateSince.Format(time.RFC3339)
	return fmt.Sprintf(`{"stateType": "%s", "stateSince": "%s", "errorTrace": %s}`, s.StateType, timeStr, s.ErrorTrack)
}

func INITIALIZINGState() State {
	return State{INITIALIZING, time.Now(), ErrorTrack{}}
}

func OKState() State {
	return State{OK, time.Now(), ErrorTrack{}}
}

func INTERRUPTEDState(errorType string, message string) State {
	return State{INTERRUPTED, time.Now(), ErrorTrack{errorType, message}}
}

func ErrorOFFState(errorType string, message string) State {
	return State{OFF, time.Now(), ErrorTrack{errorType, message}}
}

func NormalOFFState() State {
	return State{OFF, time.Now(), ErrorTrack{}}
}

type DataUpdater interface {
	// Init overwrites the storage with a set of items for each collection, if the new version > the old one
	// If the underlying data storage returns an error during this operation, the SDK will take it, log it,
	// and set the data source state to INTERRUPTED. It will not return the error to other level,
	// but will simply return false to indicate that the operation failed.
	Init(allDate map[Category]map[string]Item, version int64) bool

	// Upsert updates or inserts an item in the specified collection. For updates, the object will only be
	// updated if the existing version is less than the new version; for inserts, if the version > the existing one, it will replace
	// the existing one. If the underlying data storage returns an error during this operation, the SDK will catch it, log it,
	// and set the state to INTERRUPTED.It will not return the error to other level,
	// but will simply return false to indicate that the operation failed.
	Upsert(category Category, key string, item Item, version int64) bool

	// StorageInitialized return true if the DataStorage is well initialized
	StorageInitialized() bool

	// GetVersion returns the latest version of storage
	GetVersion() int64

	// UpdateStatus informs the SDK of a change in the DataSynchronizer status.
	// DataSynchronizer implementations should use this method,
	// if they have any concept of being in a valid state, a temporarily disconnected state, or a permanently stopped state.
	// If the new state is different from the previous state, and/or the new error is not empty,
	// SDK will start returning the new status (adding a timestamp for the change).
	// A special case is that if the new state is INTERRUPTED,
	// but the previous state was INITIALIZING, the state will remain at INITIALIZING,
	// because INTERRUPTED is only meaningful after a successful startup.
	UpdateStatus(state State)
}

type DataUpdateStatusProvider interface {
	io.Closer
	// GetCurrentState returns the current status of the DataSynchronizer
	// All of the DataSynchronizer implementations are guaranteed to update this status
	// whenever they successfully initialize, encounter an error, or recover after an error.
	// For a custom implementation, it is the responsibility of the DataSynchronizer to report its status via DataUpdater,
	// if it does not do so, the status will always be reported as INITIALIZING.
	GetCurrentState() State

	// WaitFor waits for a desired state after bootstrapping
	// If the current state is already desired State when this function is called, it immediately returns.
	// Otherwise, it blocks until 1. the state has become desired State, 2. the state has become OFF, 3. the specified timeout elapses.
	// A scenario in which this might be useful is if you want to use SDK without waiting
	// for it to finish initialization, and then wait for initialization at a later time or on a different point.
	WaitFor(state StateType, timeout time.Duration) bool

	// WaitForOKState alias of WaitFor in OK state
	WaitForOKState(timeout time.Duration) bool
}
