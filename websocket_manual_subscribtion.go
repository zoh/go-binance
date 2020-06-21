package binance

import (
	"errors"
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

/*
 Live Subscribing/Unsubscribing to streams
*/

type StreamEventType string

const (
	SUBSCRIBE          = "SUBSCRIBE"
	UNSUBSCRIBE        = "UNSUBSCRIBE"
	LIST_SUBSCRIPTIONS = "LIST_SUBSCRIPTIONS"

	// useless now
	SET_PROPERTY = "SET_PROPERTY"
	GET_PROPERTY = "GET_PROPERTY"

	streamBinanceWs = "wss://stream.binance.com:9443/stream"
	streamBinanceFutureWs = "wss://fstream.binance.com/stream"

	// Stream Events
	StreamEventDepth       StreamEventType = "depth"
	StreamEventDepthUpdate StreamEventType = "depthUpdate"
	StreamEventAggTrade    StreamEventType = "aggTrade"
	StreamEventTrade       StreamEventType = "trade"
)

type StreamEvent struct {
	Type StreamEventType
	Data interface{}
}

type WebsocketMarketStreams struct {
	Logger *log.Logger

	DoneC chan struct{}
	c     *websocket.Conn
	m     sync.RWMutex

	Debug   bool
	reqID   uint64
	stopped chan struct{}

	intern chan respSubscribe

	streamEvents chan StreamEvent
}

func (wms *WebsocketMarketStreams) GetStreamEvents() <-chan StreamEvent {
	return wms.streamEvents
}

func (wms *WebsocketMarketStreams) getReqID() uint64 {
	wms.reqID++
	return wms.reqID
}

func (wms *WebsocketMarketStreams) debug(format string, v ...interface{}) {
	if wms.Debug {
		wms.Logger.Printf(format, v...)
	}
}

func NewMarketStreams(errHandler ErrHandler, future bool) (*WebsocketMarketStreams, error) {
	endpoint := streamBinanceWs
	if future {
		endpoint = streamBinanceFutureWs
	}

	c, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		return nil, err
	}
	//

	doneC := make(chan struct{})
	//
	ms := &WebsocketMarketStreams{
		DoneC:        doneC,
		c:            c,
		m:            sync.RWMutex{},
		Debug:        false,
		reqID:        0,
		stopped:      make(chan struct{}),
		intern:       make(chan respSubscribe),
		streamEvents: make(chan StreamEvent, 1000),
		Logger:       log.New(os.Stderr, "Binance-market-stream ", log.LstdFlags),
	}

	var handler = func(message []byte) {
		ms.debug(string(message))

		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}

		// todo: need parse stream
		stream := j.Get("stream")
		data := j.Get("data")

		if stream.Interface() != nil && data.Interface() != nil {
			// it's stream and data
			event := data.Get("e").MustString()

			switch StreamEventType(event) {
			case StreamEventDepthUpdate:
				data := ParseWsDepthEvent(j)
				ms.streamEvents <- StreamEvent{
					Type: StreamEventDepthUpdate,
					Data: data,
				}
			case "aggTrade":
				ms.streamEvents <- StreamEvent{
					Type: StreamEventAggTrade,
					Data: ParseWsAggTradeEvent(data),
				}
			case "trade":
				ms.streamEvents <- StreamEvent{
					Type: StreamEventTrade,
					Data: ParseWsTradeEvent(data),
				}
			case "kline":
			case "24hrMiniTicker":
			case "24hrTicker":

			default:
				ms.streamEvents <- StreamEvent{
					Type: StreamEventDepth,
					Data: ParseWsWsPartialDepthEvent(j),
				}

				// implement=>
				// todo: !ticker@arr
				// todo: miniticker
				//  @bookTicker
			}
		} else {
			respID, err := j.Get("id").Uint64()
			if err != nil {
				errHandler(err)
				return
			}
			// attemp, is it sync operation
			ms.intern <- respSubscribe{
				ID:     respID,
				Result: j.Get("result"),
			}
		}
	}

	go func() {
		defer func() {
			cerr := c.Close()
			if cerr != nil {
				errHandler(cerr)
			}
		}()
		defer close(doneC)
		defer close(ms.stopped)
		defer close(ms.streamEvents)
		//
		if WebsocketKeepalive {
			keepAlive(c, WebsocketTimeout)
		}

		for {
			select {
			case <-ms.stopped:
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					errHandler(err)
					return
				}
				handler(message)
			}
		}
	}()

	return ms, nil
}

// Close  websocket
func (wms *WebsocketMarketStreams) Close() error {
	return wms.c.Close()
}

// Subscribe send subscribe request
func (wms *WebsocketMarketStreams) Subscribe(params ...RequestSubscriber) error {
	if len(params) == 0 {
		return errors.New("empty params")
	}
	id := wms.getReqID()

	var list []string
	for _, p := range params {
		list = append(list, p.String())
	}

	err := wms.c.WriteJSON(reqSubscribe{
		Method: SUBSCRIBE,
		ID:     id,
		Params: list,
	})
	if err != nil {
		return err
	}

	// нужно всё равно получить то то что нам нужно
	resp := <-wms.intern
	if resp.ID != id {
		return errors.New("error consistent of req id")
	}

	return nil
}

func (wms *WebsocketMarketStreams) Unsubscribe(params ...RequestSubscriber) error {
	if len(params) == 0 {
		return errors.New("empty params")
	}
	id := wms.getReqID()

	var list []string
	for _, p := range params {
		list = append(list, p.String())
	}

	err := wms.c.WriteJSON(reqSubscribe{
		Method: UNSUBSCRIBE,
		ID:     id,
		Params: list,
	})
	if err != nil {
		return err
	}

	// нужно всё равно получить то то что нам нужно
	resp := <-wms.intern
	if resp.ID != id {
		return errors.New("error consistent of req id")
	}

	return nil
}

//
func (wms *WebsocketMarketStreams) GetSubscriptions() ([]string, error) {
	id := wms.getReqID()
	err := wms.c.WriteJSON(reqSubscribe{
		Method: LIST_SUBSCRIPTIONS,
		ID:     id,
	})
	if err != nil {
		return nil, err
	}
	// wait result
	resp := <-wms.intern
	if resp.ID != id {
		return nil, errors.New("error consistent of req id")
	}

	var res []string
	sl, _ := resp.Result.Array()
	for _, v := range sl {
		if s, ok := v.(string); ok {
			res = append(res, s)
		}
	}
	return res, nil
}

type reqSubscribe struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	ID     uint64      `json:"id"`
}

type respSubscribe struct {
	ID     uint64           `json:"id"`
	Result *simplejson.Json `json:"result"`
}

type RequestSubscriber interface {
	String() string
}

type depthSubscribeOptions struct {
	symbol string
	levels int
	speed  Speed
}

func (d depthSubscribeOptions) String() string {
	res := strings.ToLower(d.symbol) + "@depth"
	if d.levels > 0 {
		res += strconv.Itoa(d.levels)
	}
	if !d.speed.IsEmpty() {
		res += "@" + d.speed.String()
	}
	return res
}

func (d *depthSubscribeOptions) Levels(levels int) *depthSubscribeOptions {
	d.levels = levels
	return d
}

// Update Speed: 250ms, 500ms or 100ms
func (d *depthSubscribeOptions) Speed(speed Speed) *depthSubscribeOptions {
	d.speed = speed
	return d
}

func NewDepthSubscribeOptions(symbol string) *depthSubscribeOptions {
	return &depthSubscribeOptions{symbol: symbol}
}

type customSubscribeOptions struct {
	val string
}

func (c customSubscribeOptions) String() string {
	return c.val
}
func CustomSubscribeOptions(data string) *customSubscribeOptions {
	return &customSubscribeOptions{val: data}
}
