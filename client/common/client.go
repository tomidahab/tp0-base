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

type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	MaxBatch      int
}

type Client struct {
	config  ClientConfig
	conn    net.Conn
	bet     bet.Bet
	stopped bool
}

func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config, stopped: false, bet: LoadBetFromEnv(),
	}
	return client
}

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


func (c *Client) StartClientLoop() {
	
	fileName := fmt.Sprintf("agency-%s.csv", c.config.ID)

	fileContent, err := os.ReadFile(fileName)
	if err != nil {
		log.Errorf("action: read_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		os.Exit(1)
	}

	if c.stopped == true {
		return
	}

	if err := c.createClientSocket(); err != nil {
		log.Errorf("action: create_connection | result: fail | client_id: %v | error: %v", c.config.ID, err)
		os.Exit(1) 
	}


	if c.stopped == true {
		return
	}

	if err := bet.ProcessFile(c.conn, c.config.ID, string(fileContent), c.config.MaxBatch, c.config.LoopPeriod); err != nil { 
		log.Errorf("action: process_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		c.conn.Close()
		os.Exit(1)
	}

	c.conn.Close()

	if c.stopped == true {
		return
	}


	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	time.Sleep(5 * time.Second)
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