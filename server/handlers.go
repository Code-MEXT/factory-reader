package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Code-MEXT/factory-reader/db"
)

type ConnectionRequest struct {
	Name         string `json:"name"`
	Protocol     string `json:"protocol"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Topic        string `json:"topic"`
	NodeID       string `json:"node_id"`
	UnitID       int    `json:"unit_id"`
	Rack         int    `json:"rack"`
	Slot         int    `json:"slot"`
	DBNumber     int    `json:"db_number"`
	StartAddress int    `json:"start_address"`
	Quantity     int    `json:"quantity"`
}

type Connection struct {
	ID int `json:"id"`
	ConnectionRequest
}

func handleListConnections(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := database.Pool.Query(context.Background(),
			`SELECT id, name, protocol, host, port, topic, node_id, unit_id, rack, slot, db_number, start_address, quantity FROM connections ORDER BY id`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var conns []Connection
		for rows.Next() {
			var c Connection
			if err := rows.Scan(&c.ID, &c.Name, &c.Protocol, &c.Host, &c.Port, &c.Topic, &c.NodeID, &c.UnitID, &c.Rack, &c.Slot, &c.DBNumber, &c.StartAddress, &c.Quantity); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			conns = append(conns, c)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(conns)
	}
}

func handleCreateConnection(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ConnectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.Protocol == "" || req.Host == "" || req.Port == 0 {
			http.Error(w, "protocol, host, and port are required", http.StatusBadRequest)
			return
		}

		var id int
		err := database.Pool.QueryRow(context.Background(),
			`INSERT INTO connections (name, protocol, host, port, topic, node_id, unit_id, rack, slot, db_number, start_address, quantity)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
			req.Name, req.Protocol, req.Host, req.Port, req.Topic, req.NodeID, req.UnitID, req.Rack, req.Slot, req.DBNumber, req.StartAddress, req.Quantity,
		).Scan(&id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": id})
	}
}

func handleDeleteConnection(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		_, err := database.Pool.Exec(context.Background(), `DELETE FROM connections WHERE id = $1`, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleWS(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("websocket upgrade error: %v", err)
			return
		}
		hub.Add(conn)
		defer hub.Remove(conn)

		// Keep connection alive, read messages (commands from UI)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}
