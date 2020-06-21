package binance

import (
	"encoding/json"
	"github.com/bitly/go-simplejson"
	"strings"
)

func ParseWsWsPartialDepthEvent(j *simplejson.Json) WsPartialDepthEvent {
	var event WsPartialDepthEvent
	//
	stream := j.Get("stream").MustString()
	symbol := strings.Split(stream, "@")[0]
	event.Symbol = strings.ToUpper(symbol)
	data := j.Get("data").MustMap()
	event.LastUpdateID, _ = data["lastUpdateId"].(json.Number).Int64()
	bidsLen := len(data["bids"].([]interface{}))
	event.Bids = make([]Bid, bidsLen)
	for i := 0; i < bidsLen; i++ {
		item := data["bids"].([]interface{})[i].([]interface{})
		event.Bids[i] = Bid{
			Price:    item[0].(string),
			Quantity: item[1].(string),
		}
	}
	asksLen := len(data["asks"].([]interface{}))
	event.Asks = make([]Ask, asksLen)
	for i := 0; i < asksLen; i++ {

		item := data["asks"].([]interface{})[i].([]interface{})
		event.Asks[i] = Ask{
			Price:    item[0].(string),
			Quantity: item[1].(string),
		}
	}

	return event
}

func ParseWsDepthEvent(j *simplejson.Json) WsDepthEvent {
	var event WsDepthEvent
	j = j.Get("data")
	//
	event.Event = j.Get("e").MustString()
	event.Time = j.Get("E").MustInt64()
	event.Symbol = j.Get("s").MustString()
	event.UpdateID = j.Get("u").MustInt64()
	event.FirstUpdateID = j.Get("U").MustInt64()
	bidsLen := len(j.Get("b").MustArray())
	event.Bids = make([]Bid, bidsLen)
	for i := 0; i < bidsLen; i++ {
		item := j.Get("b").GetIndex(i)
		event.Bids[i] = Bid{
			Price:    item.GetIndex(0).MustString(),
			Quantity: item.GetIndex(1).MustString(),
		}
	}
	asksLen := len(j.Get("a").MustArray())
	event.Asks = make([]Ask, asksLen)
	for i := 0; i < asksLen; i++ {
		item := j.Get("a").GetIndex(i)
		event.Asks[i] = Ask{
			Price:    item.GetIndex(0).MustString(),
			Quantity: item.GetIndex(1).MustString(),
		}
	}
	return event
}

func ParseWsAggTradeEvent(j *simplejson.Json) WsAggTradeEvent {
	var event WsAggTradeEvent

	event.Event = j.Get("e").MustString()
	event.Time = j.Get("E").MustInt64()
	event.Symbol = j.Get("s").MustString()
	event.AggTradeID = j.Get("a").MustInt64()
	event.Price = j.Get("p").MustString()
	event.Quantity = j.Get("q").MustString()
	event.FirstBreakdownTradeID = j.Get("f").MustInt64()
	event.LastBreakdownTradeID = j.Get("l").MustInt64()
	event.TradeTime = j.Get("T").MustInt64()
	event.IsBuyerMaker = j.Get("m").MustBool()
	event.Placeholder = j.Get("M").MustBool()

	return event
}



func ParseWsTradeEvent(j *simplejson.Json) WsTradeEvent {
	var event WsTradeEvent

	event.Event = j.Get("e").MustString()
	event.Time = j.Get("E").MustInt64()
	event.Symbol = j.Get("s").MustString()
	event.TradeID = j.Get("t").MustInt64()
	event.Price = j.Get("p").MustString()
	event.Quantity = j.Get("q").MustString()
	event.BuyerOrderID = j.Get("b").MustInt64()
	event.SellerOrderID = j.Get("a").MustInt64()
	event.TradeTime = j.Get("T").MustInt64()
	event.IsBuyerMaker = j.Get("m").MustBool()
	event.Placeholder = j.Get("M").MustBool()

	return event
}
