package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Code-MEXT/factory-reader/db"
	"github.com/Code-MEXT/factory-reader/reader"
)

type Server struct {
	db      *db.DB
	hub     *Hub
	monitor *Monitor
	mux     *http.ServeMux
	addr    string
}

func New(addr string, database *db.DB) *Server {
	hub := NewHub()
	s := &Server{
		db:      database,
		hub:     hub,
		monitor: NewMonitor(hub),
		mux:     http.NewServeMux(),
		addr:    addr,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /api/connections", handleListConnections(s.db))
	s.mux.HandleFunc("POST /api/connections", handleCreateConnection(s.db))
	s.mux.HandleFunc("DELETE /api/connections/{id}", handleDeleteConnection(s.db))
	s.mux.HandleFunc("POST /api/test/{id}", s.handleTest)
	s.mux.HandleFunc("POST /api/monitor/{id}", s.handleStartMonitor)
	s.mux.HandleFunc("DELETE /api/monitor/{id}", s.handleStopMonitor)
	s.mux.HandleFunc("GET /api/monitor", s.handleListMonitors)
	s.mux.HandleFunc("GET /ws", handleWS(s.hub))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func (s *Server) handleTest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var c Connection
	err := s.db.Pool.QueryRow(context.Background(),
		`SELECT id, name, protocol, host, port, topic, node_id, unit_id, rack, slot, db_number, start_address, quantity FROM connections WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.Protocol, &c.Host, &c.Port, &c.Topic, &c.NodeID, &c.UnitID, &c.Rack, &c.Slot, &c.DBNumber, &c.StartAddress, &c.Quantity)
	if err != nil {
		http.Error(w, "connection not found", http.StatusNotFound)
		return
	}

	rdr := s.createReader(c)
	if rdr == nil {
		http.Error(w, "unsupported protocol", http.StatusBadRequest)
		return
	}

	// Test connection
	if err := rdr.Connect(); err != nil {
		res := &reader.Result{
			ConnectionID: c.ID,
			Name:         c.Name,
			Protocol:     c.Protocol,
			Host:         c.Host,
			Port:         c.Port,
			Connected:    false,
			Error:        err.Error(),
			Timestamp:    time.Now(),
		}
		s.hub.Broadcast(res)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
		return
	}
	defer rdr.Close()

	res, _ := rdr.Read()
	res.ConnectionID = c.ID
	res.Name = c.Name
	s.hub.Broadcast(res)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) createReader(c Connection) reader.Reader {
	switch c.Protocol {
	case "mqtt":
		return reader.NewMQTTReader(c.Host, c.Port, c.Topic)
	case "opcua":
		return reader.NewOPCUAReader(c.Host, c.Port, c.NodeID)
	case "modbus":
		return reader.NewModbusReader(c.Host, c.Port, byte(c.UnitID), uint16(c.StartAddress), uint16(c.Quantity))
	case "s7":
		return reader.NewS7Reader(c.Host, c.Port, c.Rack, c.Slot, c.DBNumber, c.StartAddress, c.Quantity)
	default:
		return nil
	}
}

func (s *Server) handleStartMonitor(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var c Connection
	err := s.db.Pool.QueryRow(context.Background(),
		`SELECT id, name, protocol, host, port, topic, node_id, unit_id, rack, slot, db_number, start_address, quantity FROM connections WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.Protocol, &c.Host, &c.Port, &c.Topic, &c.NodeID, &c.UnitID, &c.Rack, &c.Slot, &c.DBNumber, &c.StartAddress, &c.Quantity)
	if err != nil {
		http.Error(w, "connection not found", http.StatusNotFound)
		return
	}

	if !s.monitor.Start(c, s.createReader) {
		http.Error(w, "already monitoring", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (s *Server) handleStopMonitor(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var connID int
	fmt.Sscan(id, &connID)

	if !s.monitor.Stop(connID) {
		http.Error(w, "not monitoring", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (s *Server) handleListMonitors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.monitor.ActiveIDs())
}

func (s *Server) Start() error {
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("server starting on %s", s.addr)
	fmt.Printf("Open http://localhost%s in your browser\n", s.addr)
	return srv.ListenAndServe()
}
