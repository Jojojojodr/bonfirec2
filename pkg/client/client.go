package client

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var input = make(chan string)

type Client struct {
	Port    string
	Address string
	Conn    net.Conn
}

func (c *Client) Connect() {
	address := c.Address + ":" + c.Port
	go c.handleMessageInput()

	for {
		var conn net.Conn
		var err error
		conn, err = net.Dial("tcp", address)
		if err != nil {
			log.Printf("Failed to connect to server: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Connected to server at %s", address)

		// Handle communication with the server
		c.Conn = conn
		if err := c.handleConnection(); err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("Connection closed by server. Reconnecting...")
			} else {
				log.Printf("Connection ended: %v. Reconnecting...", err)
			}
		}
		conn.Close()
	}
}

func (c *Client) Close() {
	if c.Conn != nil {
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
}

func (c *Client) handleConnection() error {
	log.Printf("Connection established to server: %s", c.Conn.RemoteAddr())
	readErr := make(chan error, 1)
	go c.handleMessageRead(readErr)
	return c.handleMessageSend(readErr)
}

func (c *Client) handleMessageInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		log.Print("Enter message to send (or 'exit' to quit): ")
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		input <- message
	}
}

func (c *Client) handleMessageRead(readErr chan<- error) {
	reader := bufio.NewReader(c.Conn)
	for {
		response, err := reader.ReadString('\n')
		if err != nil {
			select {
			case readErr <- err:
			default:
			}
			return
		}
		log.Printf("Received from server: %s", strings.TrimSpace(response))
	}
}

func (c *Client) handleMessageSend(readErr <-chan error) error {
	for {
		select {
		case err := <-readErr:
			return err
		case message, ok := <-input:
			if !ok {
				return errors.New("input channel closed")
			}

			if message == "exit" {
				log.Println("Exiting...")
				c.Close()
				os.Exit(0)
			}

			if _, err := c.Conn.Write([]byte(message + "\n")); err != nil {
				return err
			}
		}
	}
}

func NewClient(port string, address string) *Client {
	return &Client{
		Port:    port,
		Address: address,
		Conn:    nil,
	}
}
