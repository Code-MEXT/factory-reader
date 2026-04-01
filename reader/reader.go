package reader

import "time"

// Result holds the outcome of a read attempt.
type Result struct {
	ConnectionID int       `json:"connection_id,omitempty"`
	Name         string    `json:"name,omitempty"`
	Protocol     string    `json:"protocol"`
	Host         string    `json:"host"`
	Port         int       `json:"port"`
	Connected    bool      `json:"connected"`
	Data         any       `json:"data,omitempty"`
	Error        string    `json:"error,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// Reader is the interface all protocol readers implement.
type Reader interface {
	// Connect tests connectivity to the target.
	Connect() error
	// Read performs a single read and returns the result.
	Read() (*Result, error)
	// Close cleans up resources.
	Close() error
}
