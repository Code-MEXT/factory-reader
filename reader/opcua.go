package reader

import (
	"context"
	"fmt"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

type OPCUAReader struct {
	Host   string
	Port   int
	NodeID string
	client *opcua.Client
}

func NewOPCUAReader(host string, port int, nodeID string) *OPCUAReader {
	return &OPCUAReader{Host: host, Port: port, NodeID: nodeID}
}

func (r *OPCUAReader) Connect() error {
	endpoint := fmt.Sprintf("opc.tcp://%s:%d", r.Host, r.Port)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, err := opcua.NewClient(endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
	if err != nil {
		return fmt.Errorf("opcua client create error: %w", err)
	}
	if err := c.Connect(ctx); err != nil {
		return fmt.Errorf("opcua connect error: %w", err)
	}
	r.client = c
	return nil
}

func (r *OPCUAReader) Read() (*Result, error) {
	res := &Result{
		Protocol:  "opcua",
		Host:      r.Host,
		Port:      r.Port,
		Connected: r.client != nil,
		Timestamp: time.Now(),
	}
	if r.client == nil {
		res.Error = "not connected"
		return res, fmt.Errorf("opcua not connected")
	}

	nodeID, err := ua.ParseNodeID(r.NodeID)
	if err != nil {
		res.Error = fmt.Sprintf("invalid node id: %v", err)
		return res, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &ua.ReadRequest{
		NodesToRead: []*ua.ReadValueID{
			{NodeID: nodeID, AttributeID: ua.AttributeIDValue},
		},
	}
	resp, err := r.client.Read(ctx, req)
	if err != nil {
		res.Error = fmt.Sprintf("read error: %v", err)
		return res, err
	}
	if len(resp.Results) > 0 {
		res.Data = resp.Results[0].Value.Value()
	}
	return res, nil
}

func (r *OPCUAReader) Close() error {
	if r.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return r.client.Close(ctx)
	}
	return nil
}
