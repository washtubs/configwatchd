package configwatchd

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

const socketPortDefault int = 53673

// Exposed for RPC communication
type NoopResponse struct {
}

// Exposed for RPC communication
type NoopRequest struct {
}

type Receiver struct {
	q             *queue
	queueDisabled bool
}

type FlushOpts struct {
	// Specific keys to be executed. If empty, the entire queue is flushed
	Keys []string
	// Don't execute, simply clear the queue, or the selected items
	Clear bool
}

// Processes queued items, either just clearing or executing them
func (r *Receiver) Flush(opts FlushOpts, resp *NoopResponse) error {
	if r.queueDisabled {
		return errors.New("Queue is disabled, nothing to do.")
	}
	if len(opts.Keys) > 0 {
		if !opts.Clear {
			r.q.execute(opts.Keys)
		} else {
			r.q.clear(opts.Keys)
		}
	} else {
		if !opts.Clear {
			r.q.executeAll()
		} else {
			r.q.clearAll()
		}
	}
	return nil
}

type ListResponse struct {
	ConfigKeys []string
}

// Gives a listing of items that are queued
func (r *Receiver) List(opts NoopRequest, resp *ListResponse) error {
	if r.queueDisabled {
		return errors.New("Queue is disabled, nothing to do.")
	}
	resp.ConfigKeys = r.q.list()
	return nil
}

func setupReceiver(q *queue, opts ServerOptions) (net.Listener, error) {
	rcv := &Receiver{
		q:             q,
		queueDisabled: !opts.Queue,
	}
	srv := rpc.NewServer()
	err := srv.Register(rcv)
	if err != nil {
		return nil, err
	}

	srv.HandleHTTP("configwatchd", "configwatchd-debug")

	l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(socketPortDefault))

	if err != nil {
		panic("Couldn't serve: " + err.Error())
	}

	go func() {
		err := http.Serve(l, nil)
		if err != nil {
			log.Printf("HTTP server failed: %s", err.Error())
		}

	}()

	return l, nil

}

type Client struct {
	*rpc.Client
}

func (c *Client) flush(opts FlushOpts) error {
	response := NoopResponse{}
	return c.Call("Receiver.Flush", opts, &response)
}

func (c *Client) list() ([]string, error) {
	response := ListResponse{}
	err := c.Call("Receiver.List", NoopRequest{}, &response)
	return response.ConfigKeys, err
}

func newClient() (*Client, error) {
	socketAddr := fmt.Sprintf("localhost:%d", socketPortDefault)
	client, err := rpc.DialHTTPPath("tcp", socketAddr, "configwatchd")
	if err != nil {
		return nil, fmt.Errorf("Error dialing %s: %w", socketAddr, err)
	}
	return &Client{client}, nil
}
