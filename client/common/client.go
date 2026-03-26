package common

import (
	"bufio"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID             string
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	BatchMaxAmount int
	DataPath       string
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
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	var AllBetsSent bool = false
	currentDelay := c.config.LoopPeriod

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	br, err := NewBettingReader(c.config.DataPath)
	if err != nil {
		log.Errorf("action: open_file | result: fail | error: %v", err)
		return
	}
	defer br.Close()

	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		select {
		case <-sigs:
			log.Infof("action: sigterm_received | result: success | client_id: %v", c.config.ID)
			if c.conn != nil {
				c.conn.Close()
			}
			return
		default:
		}

		if AllBetsSent || c.conn == nil {
			err := c.createClientSocket()
			if err != nil {
				time.Sleep(currentDelay)
				continue
			}
		}

		if AllBetsSent {
			SendQueryMessage(c.conn, c.config.ID)
			log.Infof("action: send_query_message | result: success | client_id: %v", c.config.ID)

			resp, err := bufio.NewReader(c.conn).ReadString('\n')
			c.conn.Close()
			c.conn = nil

			if err == nil && strings.Contains(resp, "FINISHED") {
				log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
				return
			}

			time.Sleep(currentDelay)
			currentDelay *= 2
			if currentDelay > 60*time.Second {
				currentDelay = 60 * time.Second
			}
			continue
		}

		sentCount, sendErr := SendBets(c.conn, br, c.config.BatchMaxAmount, c.config.ID)

		if sentCount > 0 {
			_, err = bufio.NewReader(c.conn).ReadString('\n')
			log.Infof("action: receive_ack | result: success | sent_count: %v", sentCount)
			if err != nil {
				log.Errorf("action: receive_ack_aca | error: %v", err)
				c.conn.Close()
				c.conn = nil
			}
		}

		if sendErr != nil {
			if sendErr.Error() == "EOF" {
				AllBetsSent = true
				SendFinishMessage(c.conn, c.config.ID)
				_, ackErr := bufio.NewReader(c.conn).ReadString('\n')
				if ackErr != nil {
					log.Errorf("action: receive_finish_ack | error: %v", ackErr)
				}

				log.Infof("action: send_finish_message | result: success | client_id: %v", c.config.ID)
				c.conn.Close()
				c.conn = nil
				currentDelay = c.config.LoopPeriod
			} else {
				log.Errorf("action: send_data | result: fail | error: %v", sendErr)
				c.conn.Close()
				c.conn = nil
			}
		}

		time.Sleep(c.config.LoopPeriod)
	}
}
