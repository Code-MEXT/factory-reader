package server

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Code-MEXT/factory-reader/reader"
)

type Monitor struct {
	mu       sync.RWMutex
	active   map[int]context.CancelFunc // connection ID -> cancel
	hub      *Hub
}

func NewMonitor(hub *Hub) *Monitor {
	return &Monitor{
		active: make(map[int]context.CancelFunc),
		hub:    hub,
	}
}

func (m *Monitor) IsActive(id int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.active[id]
	return ok
}

func (m *Monitor) ActiveIDs() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]int, 0, len(m.active))
	for id := range m.active {
		ids = append(ids, id)
	}
	return ids
}

func (m *Monitor) Start(conn Connection, createReader func(Connection) reader.Reader) bool {
	m.mu.Lock()
	if _, exists := m.active[conn.ID]; exists {
		m.mu.Unlock()
		return false // already monitoring
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.active[conn.ID] = cancel
	m.mu.Unlock()

	go m.loop(ctx, conn, createReader)
	return true
}

func (m *Monitor) Stop(id int) bool {
	m.mu.Lock()
	cancel, ok := m.active[id]
	if ok {
		cancel()
		delete(m.active, id)
	}
	m.mu.Unlock()
	return ok
}

func (m *Monitor) loop(ctx context.Context, conn Connection, createReader func(Connection) reader.Reader) {
	defer func() {
		m.mu.Lock()
		delete(m.active, conn.ID)
		m.mu.Unlock()
	}()

	rdr := createReader(conn)
	if rdr == nil {
		return
	}

	if err := rdr.Connect(); err != nil {
		res := &reader.Result{
			ConnectionID: conn.ID,
			Name:         conn.Name,
			Protocol:     conn.Protocol,
			Host:         conn.Host,
			Port:         conn.Port,
			Connected:    false,
			Error:        err.Error(),
			Timestamp:    time.Now(),
		}
		m.hub.Broadcast(res)
		return
	}
	defer rdr.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Do an immediate first read
	m.readAndBroadcast(rdr, conn)

	for {
		select {
		case <-ctx.Done():
			log.Printf("monitor stopped for connection %d (%s)", conn.ID, conn.Name)
			return
		case <-ticker.C:
			m.readAndBroadcast(rdr, conn)
		}
	}
}

func (m *Monitor) readAndBroadcast(rdr reader.Reader, conn Connection) {
	res, err := rdr.Read()
	if err != nil {
		res = &reader.Result{
			Protocol:  conn.Protocol,
			Host:      conn.Host,
			Port:      conn.Port,
			Connected: false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}
	res.ConnectionID = conn.ID
	res.Name = conn.Name
	m.hub.Broadcast(res)
}