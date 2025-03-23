package common

import (
	"bufio"
	"fmt"
	"net"
	"time"
	"os/signal"
	"os"
	"syscall"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config, stopped: false
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {

	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {

		if c.stopped == true {
			return
		}
		// Create the connection the server in every loop iteration. Send an
		if err := c.createClientSocket(); err != nil {
			log.Errorf("action: create_connection | result: fail | client_id: %v | error: %v", c.config.ID, err)
			os.Exit(1) 
		}

		if c.stopped == true {
			return
		}

		// TODO: Modify the send to avoid short-write
		fmt.Fprintf(
			c.conn,
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()

		if c.stopped == true {
			return
		}

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			os.Exit(1)
		}

		if c.stopped == true {
			return
		}

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	os.Exit(0)
}


func (c *Client) Close() {
    log.Infof("action: shutdown | result: in_progress | client_id: %v", c.config.ID)
    c.stopped = true
    if c.conn != nil {
        c.conn.Close()
        c.conn = nil
    }
    log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
}
