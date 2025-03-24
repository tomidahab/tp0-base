package common

import (
	"net"
	"time"
	"os"
	"fmt"

	"github.com/op/go-logging"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/common/bet"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	MaxBatch      int
}

// Client Entity that encapsulates how
type Client struct {
	config  ClientConfig
	conn    net.Conn
	bet     bet.Bet
	stopped bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config, stopped: false, bet: LoadBetFromEnv(),
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
/*func (c *Client) StartClientLoop() {
	
	fileName := fmt.Sprintf("agency-%s.csv", c.config.ID)

	// Read the file content
	fileContent, err := os.ReadFile(fileName)
	if err != nil {
		log.Errorf("action: read_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		os.Exit(1)
	}
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {

		if c.stopped == true {
			return
		}
		// Create the connection the server in every loop iteration. Send an

		// log.Infof("DELETE about to create conn")

		if err := c.createClientSocket(); err != nil {
			log.Errorf("action: create_connection | result: fail | client_id: %v | error: %v", c.config.ID, err)
			os.Exit(1) 
		}

		// log.Infof("DELETE created conn")


		if c.stopped == true {
			return
		}

		if err := bet.ProcessFile(c.conn, c.config.ID, string(fileContent), c.config.MaxBatch); err != nil { //TODO cambiar el 100 por la config
			log.Errorf("action: process_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
			c.conn.Close()
			os.Exit(1)
		}

		c.conn.Close()

		if c.stopped == true {
			return
		}


		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	os.Exit(0)
}*/

func (c *Client) StartClientLoop() {
	
	fileName := fmt.Sprintf("agency-%s.csv", c.config.ID)

	// Read the file content
	fileContent, err := os.ReadFile(fileName)
	if err != nil {
		log.Errorf("action: read_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		os.Exit(1)
	}

	if c.stopped == true {
		return
	}
	// Create the connection the server in every loop iteration. Send an

	// log.Infof("DELETE about to create conn")

	if err := c.createClientSocket(); err != nil {
		log.Errorf("action: create_connection | result: fail | client_id: %v | error: %v", c.config.ID, err)
		os.Exit(1) 
	}

	// log.Infof("DELETE created conn")


	if c.stopped == true {
		return
	}

	if err := bet.ProcessFile(c.conn, c.config.ID, string(fileContent), c.config.MaxBatch, c.config.LoopPeriod); err != nil { //TODO cambiar el 100 por la config
		log.Errorf("action: process_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		c.conn.Close()
		os.Exit(1)
	}

	c.conn.Close()

	if c.stopped == true {
		return
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

func LoadBetFromEnv() bet.Bet {
	return bet.Bet{
		Nombre:     os.Getenv("NOMBRE"),
		Apellido:   os.Getenv("APELLIDO"),
		DNI:        os.Getenv("DOCUMENTO"),
		Nacimiento: os.Getenv("NACIMIENTO"),
		Numero:     os.Getenv("NUMERO"),
	}
}