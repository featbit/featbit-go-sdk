package datasynchronization

import (
	"encoding/json"
	"fmt"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/types/data"
	"github.com/featbit/featbit-go-sdk/internal/util"
	"github.com/featbit/featbit-go-sdk/internal/util/log"
	"github.com/gorilla/websocket"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	authParams                        = "?token=%s&type=server"
	jsonParsingErrorMsg               = "fb json format is invalid"
	r2pChFullErrStr                   = "too many sync message in queue, skip the message and restart"
	pingInterval                      = 10 * time.Second
	closeTimeOut                      = 10 * time.Second
	invalidRequestClose               = 4003
	invalidRequestCloseReason         = "invalid request"
	CloseAndThenReconnByDatasyncError = "data sync error"
)

var (
	// JsonParsingError internal error, normally internal test use
	JsonParsingError = fmt.Errorf(jsonParsingErrorMsg)
	R2PChCap         = 100
)

type syncMessage struct {
	data                  *data.All
	err                   error
	ok                    bool
	isReconnect           bool
	isNormalClose         bool
	isRequestInvalidClose bool
	isPeerAwayClose       bool
	isOtherClose          bool
	isJsonParsingErr      bool
	isOtherErr            bool
}

func newSyncMessage(bytes []byte, err error) *syncMessage {
	if err != nil {
		if e, success := err.(*websocket.CloseError); success {
			switch e.Code {
			case websocket.CloseNormalClosure:
				return &syncMessage{err: err, isNormalClose: true}
			case invalidRequestClose:
				return &syncMessage{err: err, isRequestInvalidClose: true}
			case websocket.CloseGoingAway:
				return &syncMessage{err: err, isPeerAwayClose: true, isReconnect: true}
			default:
				return &syncMessage{err: err, isOtherClose: true, isReconnect: true}

			}
		}
		return &syncMessage{err: err, isOtherErr: true, isReconnect: true}
	}
	var m data.Message
	e := json.Unmarshal(bytes, &m)
	if e != nil {
		return &syncMessage{err: JsonParsingError, isJsonParsingErr: true}
	}
	// ignore pong message
	if !m.IsSyncMessage() {
		return nil
	}
	var all data.All
	e = json.Unmarshal(bytes, &all)
	if e != nil || !all.IsProcessData() {
		return &syncMessage{err: JsonParsingError, isJsonParsingErr: true}
	}
	return &syncMessage{data: &all, ok: true}
}

type Streaming struct {
	maxRetryTimes int64
	context       Context
	dataUpdater   DataUpdater
	strategy      *BackoffAndJitterStrategy
	// start actions should call only one time
	startOnce sync.Once
	// ready actions should call only one time
	readyOnce sync.Once
	// notify that sdk client data sync is ready
	readyCh chan struct{}
	// stream ready sig
	initialized bool
	// close stream action should call only one time
	closeOnce sync.Once
	// notify that stream should clean all resources and quite
	closeCh chan struct{}
	// stream close sig
	streamClosed bool
	lock         sync.RWMutex
	// ws connected tag
	wsConnected      bool
	conn             *websocket.Conn
	connRetryCounter int64
	pingScheduler    *time.Ticker
	r2pChan          chan *syncMessage
}

func NewStreaming(context Context, dataUpdater DataUpdater, firstRetryDelay time.Duration, maxRetryTimes int) *Streaming {
	return &Streaming{
		context:       context,
		dataUpdater:   dataUpdater,
		maxRetryTimes: int64(maxRetryTimes),
		strategy:      NewWithFirstRetryDelay(firstRetryDelay),
		readyCh:       make(chan struct{}),
		closeCh:       make(chan struct{}),
	}
}

func (s *Streaming) Close() error {
	s.closeOnce.Do(func() {
		log.LogInfo("FB JAVA SDK: streaming is stopping...")
		s.streamClosed = true
		close(s.closeCh)
	})
	return nil
}

func (s *Streaming) IsInitialized() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.initialized
}

func (s *Streaming) Start() <-chan struct{} {
	s.startOnce.Do(func() {
		log.LogDebug("Streaming Starting...")
		atomic.AddInt64(&s.connRetryCounter, 0)
		s.strategy.SetGoodRunAtNow()
		go s.connectRoutine()
	})
	return s.readyCh
}

func (s *Streaming) sendMessageToServer(messageType int, msg []byte) error {
	if !s.isWsConnected() {
		return nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.conn.WriteMessage(messageType, msg)
}

func (s *Streaming) sendCloseMessageToServer(closeCode int, closeText string) error {
	return s.sendMessageToServer(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, closeText))
}

func (s *Streaming) isWsConnected() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.conn != nil && s.wsConnected
}

func (s *Streaming) noMoreReconnect() {
	s.readyOnce.Do(func() {
		close(s.readyCh)
	})
	s.lock.Lock()
	s.streamClosed = true
	s.lock.Unlock()
}

func (s *Streaming) clean() {
	if !s.isWsConnected() {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.wsConnected = false
	_ = s.conn.Close()
	if s.r2pChan != nil {
		close(s.r2pChan)
	}
}

func (s *Streaming) reconnect() {
	// release resources
	s.clean()
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.streamClosed {
		log.LogDebug("force to quit, no more reconnect")
		return
	}
	go s.connectRoutine()
}

func (s *Streaming) onClose(code int, _ string) error {
	// handle close code
	log.LogDebug("Streaming WebSocket close reason: %d", code)
	switch code {
	case websocket.CloseNormalClosure:
		s.dataUpdater.UpdateStatus(NormalOFFState())
	case invalidRequestClose:
		s.dataUpdater.UpdateStatus(ErrorOFFState(RequestInvalidError, invalidRequestCloseReason))
	case websocket.CloseGoingAway:
		// data updater has handled the data sync error, just ignore it
		break
	default:
		s.dataUpdater.UpdateStatus(ErrorOFFState(UnknownCloseCode, strconv.Itoa(code)))
	}
	return nil
}

func (s *Streaming) onPing() error {
	return s.sendMessageToServer(websocket.TextMessage, []byte(data.DefaultPingMessage))
}

func (s *Streaming) onOpen() error {
	if !s.isWsConnected() {
		return nil
	}
	log.LogDebug("Ask Data Updating")
	createJson := func(version int64) []byte {
		return []byte(fmt.Sprintf(data.DefaultSyncMessage, version))
	}
	var bytes []byte
	if s.dataUpdater.StorageInitialized() {
		bytes = createJson(s.dataUpdater.GetVersion())
	} else {
		bytes = createJson(0)
	}
	return s.sendMessageToServer(websocket.TextMessage, bytes)
}

func (s *Streaming) onDataProcess(allData *data.All) bool {
	log.LogDebug("Streaming WebSocket is processing data")
	newData := allData.Data.ToStorageType()
	var success bool = true
	switch allData.Data.EventType {
	case data.FullOp:
		success = s.dataUpdater.Init(newData, allData.Data.GetTimestamp())
	case data.PatchOp:
	LOOP:
		for cat, items := range newData {
			for key, item := range items {
				if !s.dataUpdater.Upsert(cat, key, item, item.GetTimestamp()) {
					success = false
					break LOOP
				}
			}
		}
	}
	if success {
		s.readyOnce.Do(func() {
			log.LogDebug("processing data is well done")
			s.initialized = true
			close(s.readyCh)
			s.dataUpdater.UpdateStatus(OKState())
		})
	}
	return success
}

func (s *Streaming) connectRoutine() {
	for s.connRetryCounter <= s.maxRetryTimes && !s.streamClosed {
		network := s.context.GetNetwork()
		dialer := network.GetWebsocketClient().(*websocket.Dialer)
		streamingUri := s.context.GetStreamingUri()
		token := util.BuildToken(s.context.GetEnvSecret())
		urlFormat := strings.Join([]string{streamingUri, authParams}, "")
		url := fmt.Sprintf(urlFormat, token)
		conn, resp, err := dialer.Dial(url, network.GetHeaders(nil))
		if err != nil {
			log.LogDebug("Err in connecting ws server, http code = %v", resp.StatusCode)
			s.dataUpdater.UpdateStatus(INTERRUPTEDState(NetworkError, err.Error()))
			if _, ok := err.(*net.DNSError); ok {
				log.LogError("FB GO SDK: Host unknown: %s", err.Error())
				s.noMoreReconnect()
				return
			}
			delayToReconnect := s.strategy.NextDelay()
			log.LogError("FB GO SDK: Streaming Websocket network error  : %s, try to reconnect...", err.Error())
			time.Sleep(delayToReconnect)
			continue
		}
		log.LogDebug("ws conn is done")
		s.strategy.SetGoodRunAtNow()
		s.lock.Lock()
		s.wsConnected = true
		s.conn = conn
		s.conn.SetCloseHandler(s.onClose)
		s.r2pChan = make(chan *syncMessage, R2PChCap)
		s.lock.RUnlock()
		_ = s.onOpen()
		go s.readRoutine()
		go s.dataProcessRoutine()
		return
	}
}

func (s *Streaming) readRoutine() (isConnect bool) {
	if !s.isWsConnected() {
		return
	}
	defer s.reconnect()

	for {
		_, jsonBytes, err := s.conn.ReadMessage()
		msg := newSyncMessage(jsonBytes, err)
		// ignore pong message
		if msg == nil {
			continue
		}
		// 1001 close error: data sync error, just stop routines and restart them
		if msg.isPeerAwayClose {
			return
		}
		// 1000 close code and 4003 close code, close and no more restart
		if msg.isNormalClose || msg.isRequestInvalidClose {
			s.noMoreReconnect()
			return
		}
		// json parsing error, close and no more restart; fatal error should contact to FeatBit team
		if msg.isJsonParsingErr {
			log.LogError("FB GO SDK: Streaming WebSocket Failure: json parsing error, fatal error should contact to FeatBit team")
			s.dataUpdater.UpdateStatus(ErrorOFFState(DataInvalidError, DataInvalidError))
			s.noMoreReconnect()
			return
		}
		// the return msg is ok, other close code, or other err, notify data process routine
		select {
		case s.r2pChan <- msg:
		default:
			log.LogDebug(r2pChFullErrStr)
			s.dataUpdater.UpdateStatus(INTERRUPTEDState(UnknownError, r2pChFullErrStr))
			return
		}
		// other close code, or other err, stop routines and restart them
		if err != nil {
			return
		}
	}
}

func (s *Streaming) dataProcessRoutine() {
	if !s.wsConnected {
		return
	}
	// start ping scheduler, stop it at quiting the routine
	s.pingScheduler = time.NewTicker(pingInterval)
	defer s.pingScheduler.Stop()

	// to listen data to process,
	// error to restart,
	// and close signal to quit from the streaming
	for {
		select {
		case t := <-s.pingScheduler.C:
			log.LogTrace("ping in %s", t)
			_ = s.onPing()
		case syncMsg, ok := <-s.r2pChan:
			if !ok {
				log.LogDebug("quit the routine by unexpected error or close, maybe reconnect later")
				return
			}
			if syncMsg.ok && !s.onDataProcess(syncMsg.data) {
				// data sync failed, should to reconnect to server
				_ = s.sendCloseMessageToServer(websocket.CloseGoingAway, CloseAndThenReconnByDatasyncError)
			} else if syncMsg.isOtherErr {
				// handle reconnect-able error
				log.LogWarn("FB GO SDK: Streaming WbSocket will reconnect because of %v", syncMsg.err.Error())
				s.dataUpdater.UpdateStatus(INTERRUPTEDState(WebsocketError, syncMsg.err.Error()))
			}
		case <-s.closeCh:
			log.LogDebug("force to close streaming because of SDK quit")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := s.sendCloseMessageToServer(websocket.CloseNormalClosure, "")
			if err != nil {
				log.LogError("FB GO SDK: unknown error in closing streaming, %v", err.Error())
				s.dataUpdater.UpdateStatus(ErrorOFFState(UnknownError, err.Error()))
				return
			}
			select {
			case <-s.r2pChan:
			case <-time.After(closeTimeOut):
				log.LogDebug("time out in closing streaming, force to quit")
				s.dataUpdater.UpdateStatus(ErrorOFFState(WebsocketCloseTimeout, WebsocketCloseTimeout))
			}
			return
		}
	}

}
