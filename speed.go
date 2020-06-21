package binance

import "strconv"

type Speed struct {
	value int
	ms    bool
	s     bool
}

func NewSpeed(time int, ms, s bool) Speed {
	return Speed{
		value: time,
		ms:    ms,
		s:     s,
	}
}

func (s Speed) IsEmpty() bool {
	return s.value == 0 && s.ms == false && s.s == false
}

func (s Speed) String() string {
	r := strconv.Itoa(s.value)
	if s.ms {
		r += "ms"
	} else if s.s {
		r += "s"
	}
	return r
}
