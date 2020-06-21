package binance_test

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/zoh/go-binance"
	"testing"
	"time"
)

type websocketMarketStreamsTestSuite struct {
	suite.Suite
	ms *binance.WebsocketMarketStreams
}

var errHandler = func(err error) {
	fmt.Println(err)
}

func (s *websocketMarketStreamsTestSuite) SetupTest() {
	ms, err := binance.NewMarketStreams(errHandler, false)
	s.Empty(err)
	ms.Debug = true
	s.ms = ms
	fmt.Println("create new stream")
}

func (s *websocketMarketStreamsTestSuite) TearDownTest() {
	fmt.Println("stop")
	s.ms.Close()
}

func TestUserStreamService(t *testing.T) {
	suite.Run(t, new(websocketMarketStreamsTestSuite))
}

func (s *websocketMarketStreamsTestSuite) TestConnection() {
	fmt.Println("start")
	go func() {
		time.Sleep(1 * time.Second)
		s.ms.Close()
	}()
	<-s.ms.DoneC
	fmt.Println("Ok")
}

func (s *websocketMarketStreamsTestSuite) TestSubscription() {
	resp, err := s.ms.GetSubscriptions()
	if err != nil {
		s.Error(err)
	}
	s.Equal(resp, []string(nil), "need result of empty slice []")
	//
	err = s.ms.Subscribe()
	s.Error(err, "must error empty params")

	err = s.ms.Subscribe(
		binance.NewDepthSubscribeOptions("BTCUSDT").
			Speed(binance.NewSpeed(100, true, false)).
			Levels(5),
		binance.CustomSubscribeOptions("thetausdt@depth@100ms"),
	)

	ss, _ := s.ms.GetSubscriptions()
	s.Equal(ss, []string{"btcusdt@depth5@100ms", "thetausdt@depth@100ms"})

	_ = s.ms.Unsubscribe(binance.CustomSubscribeOptions("thetausdt@depth@100ms"))
	ss, _ = s.ms.GetSubscriptions()
	s.Equal(ss, []string{"btcusdt@depth5@100ms"})
}

func (s *websocketMarketStreamsTestSuite) TestSubscriptionOnTradeAndAggTrade() {
	_ = s.ms.Subscribe(
		binance.CustomSubscribeOptions("btcusdt@aggTrade"),
		binance.CustomSubscribeOptions("btcusdt@trade"),
	)
	var done = make(chan struct{})
	s.ms.Debug = false

	go func() {
		var hasAggTrade bool
		var hasTrade bool
		for v := range s.ms.GetStreamEvents() {
			if v.Type == binance.StreamEventAggTrade {
				if _, ok := v.Data.(binance.WsAggTradeEvent); ok {
					hasAggTrade = true
				}
			}
			if v.Type == binance.StreamEventAggTrade {
				if _, ok := v.Data.(binance.WsAggTradeEvent); ok {
					hasTrade = true
				}
			}

			if hasAggTrade && hasTrade {
				fmt.Println("get all")
				break
			}
		}
		done <- struct{}{}
	}()

	<-done
}

func TestWebsocketMarketStreams_Futures(t *testing.T) {
	ms, err := binance.NewMarketStreams(errHandler, true)
	if err != nil {
		t.Fatal(err)
	}
	ms.Debug = true

	_ = ms.Subscribe(
		binance.CustomSubscribeOptions("btcusdt@depth"),
		binance.CustomSubscribeOptions("btcusdt@aggTrade"),
	)

	go func() {
		for v := range ms.GetStreamEvents() {
			fmt.Println(v, "+?>")
		}
	}()

	time.Sleep(time.Second * 3)
}
