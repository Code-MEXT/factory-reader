package reader

import (
	"fmt"
	"time"

	"github.com/robinson/gos7"
)

type S7Reader struct {
	Host         string
	Port         int
	Rack         int
	Slot         int
	DBNumber     int
	StartAddress int
	Quantity     int
	handler      *gos7.TCPClientHandler
	client       gos7.Client
	connected    bool
}

func NewS7Reader(host string, port, rack, slot, dbNumber, startAddr, quantity int) *S7Reader {
	return &S7Reader{
		Host:         host,
		Port:         port,
		Rack:         rack,
		Slot:         slot,
		DBNumber:     dbNumber,
		StartAddress: startAddr,
		Quantity:     quantity,
	}
}

func (r *S7Reader) Connect() error {
	r.handler = gos7.NewTCPClientHandler(fmt.Sprintf("%s:%d", r.Host, r.Port), r.Rack, r.Slot)
	r.handler.Timeout = 5 * time.Second
	r.handler.IdleTimeout = 5 * time.Second

	if err := r.handler.Connect(); err != nil {
		return fmt.Errorf("s7 connect error: %w", err)
	}
	r.client = gos7.NewClient(r.handler)
	r.connected = true
	return nil
}

func (r *S7Reader) Read() (*Result, error) {
	res := &Result{
		Protocol:  "s7",
		Host:      r.Host,
		Port:      r.Port,
		Connected: r.connected,
		Timestamp: time.Now(),
	}
	if !r.connected {
		res.Error = "not connected"
		return res, fmt.Errorf("s7 not connected")
	}

	buf := make([]byte, r.Quantity)
	if err := r.client.AGReadDB(r.DBNumber, r.StartAddress, r.Quantity, buf); err != nil {
		res.Connected = false
		r.connected = false
		res.Error = fmt.Sprintf("read error: %v", err)
		return res, err
	}
	res.Data = buf
	return res, nil
}

func (r *S7Reader) Close() error {
	r.connected = false
	r.handler.Close()
	return nil
}
