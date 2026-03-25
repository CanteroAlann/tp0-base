package common

import (
	"bufio"
	"encoding/binary"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

func parseResponseFieldValues(response string) (string, string, bool) {
	fields := strings.Fields(response)
	parsed := map[string]string{}

	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			continue
		}
		parsed[parts[0]] = parts[1]
	}

	dni, dniOK := parsed["dni"]
	numero, numeroOK := parsed["numero"]
	if !dniOK || !numeroOK {
		return "", "", false
	}

	return dni, numero, true
}

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
		log.Errorf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		select {
		case <-sigs:
			log.Infof("action: sigterm_received | result: success | client_id: %v", c.config.ID)
			return

		default:
		}

		for {
			err := c.createClientSocket()
			if err == nil {
				break
			}

			select {
			case <-sigs:
				log.Infof("action: sigterm_received | result: success | client_id: %v", c.config.ID)
				return
			default:
			}

			time.Sleep(c.config.LoopPeriod)
		}

		sentCount, err := SendData(c.conn, c.config.BatchMaxAmount, c.config.DataPath, c.config.ID)

		log.Infof("action: message_send | result: success | amount: %s", sentCount)

		msg := Message{
			PayloadSize: 0,
			Payload:     []byte{},
		}

		werr := binary.Write(c.conn, binary.BigEndian, msg.PayloadSize)
		if werr != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				werr,
			)
			c.conn.Close()
			return
		}

		_, err = c.conn.Write(msg.Payload)
		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			c.conn.Close()
			return
		}

		rta, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		dni, numero, ok := parseResponseFieldValues(rta)
		if !ok {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: invalid_response_format | msg: %v",
				c.config.ID,
				rta,
			)
			return
		}

		log.Infof("action: apuesta_enviada | result: success | dni: %s | numero: %s", dni, numero)

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
