package binance

import (
	"context"
	"encoding/json"
)

// CreateWithdrawService create withdraw
type CreateWithdrawService struct {
	c       *Client
	asset   string
	address string
	amount  string
	name    *string
}

// Asset set asset
func (s *CreateWithdrawService) Asset(asset string) *CreateWithdrawService {
	s.asset = asset
	return s
}

// Address set address
func (s *CreateWithdrawService) Address(address string) *CreateWithdrawService {
	s.address = address
	return s
}

// Amount set amount
func (s *CreateWithdrawService) Amount(amount string) *CreateWithdrawService {
	s.amount = amount
	return s
}

// Name set name
func (s *CreateWithdrawService) Name(name string) *CreateWithdrawService {
	s.name = &name
	return s
}

// Do send request
func (s *CreateWithdrawService) Do(ctx context.Context) (err error) {
	r := &request{
		method:   "POST",
		endpoint: "/wapi/v3/withdraw.html",
		secType:  secTypeSigned,
	}
	r.setParam("asset", s.asset)
	r.setParam("address", s.address)
	r.setParam("amount", s.amount)
	if s.name != nil {
		r.setParam("name", *s.name)
	}
	_, err = s.c.callAPI(ctx, r)
	return err
}

// ListWithdrawsService list withdraws
type ListWithdrawsService struct {
	c         *Client
	asset     *string
	status    *int
	startTime *int64
	endTime   *int64
}

// Asset sets the asset parameter.
func (s *ListWithdrawsService) Asset(asset string) *ListWithdrawsService {
	s.asset = &asset
	return s
}

// Status sets the status parameter.
func (s *ListWithdrawsService) Status(status int) *ListWithdrawsService {
	s.status = &status
	return s
}

// StartTime sets the startTime parameter.
// If present, EndTime MUST be specified. The difference between EndTime - StartTime MUST be between 0-90 days.
func (s *ListWithdrawsService) StartTime(startTime int64) *ListWithdrawsService {
	s.startTime = &startTime
	return s
}

// EndTime sets the endTime parameter.
// If present, StartTime MUST be specified. The difference between EndTime - StartTime MUST be between 0-90 days.
func (s *ListWithdrawsService) EndTime(endTime int64) *ListWithdrawsService {
	s.endTime = &endTime
	return s
}

// Do sends the request.
func (s *ListWithdrawsService) Do(ctx context.Context) (withdraws []*Withdraw, err error) {
	r := &request{
		method:   "GET",
		endpoint: "/wapi/v3/withdrawHistory.html",
		secType:  secTypeSigned,
	}
	if s.asset != nil {
		r.setParam("asset", *s.asset)
	}
	if s.status != nil {
		r.setParam("status", *s.status)
	}
	if s.startTime != nil {
		r.setParam("startTime", *s.startTime)
	}
	if s.endTime != nil {
		r.setParam("endTime", *s.endTime)
	}
	data, err := s.c.callAPI(ctx, r)
	if err != nil {
		return
	}
	res := new(WithdrawHistoryResponse)
	err = json.Unmarshal(data, res)
	if err != nil {
		return
	}
	return res.Withdraws, nil
}

// WithdrawHistoryResponse represents a response from ListWithdrawsService.
type WithdrawHistoryResponse struct {
	Withdraws []*Withdraw `json:"withdrawList"`
	Success   bool        `json:"success"`
}

// Withdraw represents a single withdraw entry.
type Withdraw struct {
	ID              string  `json:"id"`
	WithdrawOrderID string  `json:"withdrawOrderId"`
	Amount          float64 `json:"amount"`
	TransactionFee  float64 `json:"transactionFee"`
	Address         string  `json:"address"`
	AddressTag      string  `json:"addressTag"`
	TxID            string  `json:"txId"`
	Asset           string  `json:"asset"`
	ApplyTime       int64   `json:"applyTime"`
	Network         string  `json:"network"`
	Status          int     `json:"status"`
}

// GetWithdrawFeeService get withdraw fee
type GetWithdrawFeeService struct {
	c     *Client
	asset string
}

// Asset set asset
func (s *GetWithdrawFeeService) Asset(asset string) *GetWithdrawFeeService {
	s.asset = asset
	return s
}

// Do send request
func (s *GetWithdrawFeeService) Do(ctx context.Context, opts ...RequestOption) (res *WithdrawFee, err error) {
	r := &request{
		method:   "GET",
		endpoint: "/wapi/v3/withdrawFee.html",
		secType:  secTypeSigned,
	}
	r.setParam("asset", s.asset)
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = new(WithdrawFee)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// WithdrawFee withdraw fee
type WithdrawFee struct {
	Fee float64 `json:"withdrawFee"` // docs specify string value but api returns decimal
}
