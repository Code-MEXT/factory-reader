package reader

import (
	"fmt"
	"time"

	"github.com/goburrow/modbus"
)

type ModbusReader struct {
	Host         string
	Port         int
	UnitID       byte
	StartAddress uint16
	Quantity     uint16
	handler      *modbus.TCPClientHandler
	client       modbus.Client
}

func NewModbusReader(host string, port int, unitID byte, startAddr, quantity uint16) *ModbusReader {
	return &ModbusReader{
		Host:         host,
		Port:         port,
		UnitID:       unitID,
		StartAddress: startAddr,
		Quantity:     quantity,
	}
}

func (r *ModbusReader) Connect() error {
	addr := fmt.Sprintf("%s:%d", r.Host, r.Port)
	r.handler = modbus.NewTCPClientHandler(addr)
	r.handler.Timeout = 5 * time.Second
	r.handler.SlaveId = r.UnitID

	if err := r.handler.Connect(); err != nil {
		return fmt.Errorf("modbus connect error: %w", err)
	}
	r.client = modbus.NewClient(r.handler)
	return nil
}

func (r *ModbusReader) Read() (*Result, error) {
	res := &Result{
		Protocol:  "modbus",
		Host:      r.Host,
		Port:      r.Port,
		Connected: r.client != nil,
		Timestamp: time.Now(),
	}
	if r.client == nil {
		res.Error = "not connected"
		return res, fmt.Errorf("modbus not connected")
	}

	data, err := r.client.ReadHoldingRegisters(r.StartAddress, r.Quantity)
	if err != nil {
		res.Connected = false
		res.Error = fmt.Sprintf("read error: %v", err)
		return res, err
	}
	// Convert bytes to uint16 register values
	registers := make([]uint16, len(data)/2)
	for i := 0; i < len(data)-1; i += 2 {
		registers[i/2] = uint16(data[i])<<8 | uint16(data[i+1])
	}
	res.Data = registers
	return res, nil
}

func (r *ModbusReader) Close() error {
	if r.handler != nil {
		return r.handler.Close()
	}
	return nil
}
