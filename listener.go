package bonfirec2

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Jojojojodr/bonfirec2/models"

	"github.com/google/uuid"
)

var (
	Listeners          = make(map[string]*Listener)
	gruntConnections   = make(map[string]net.Conn)
	gruntConnectionsMu sync.RWMutex
)

type Listener struct {
	models.DefaultModel
	Address     string
	Port        string
	Protocol    string
	Status      string
	GruntCount  int
	LastCheckIn string
	ln          net.Listener
	quitch      chan struct{}
	msgch       chan []byte
	mu          sync.Mutex
	running     bool
}

func (l *Listener) Start() error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return errors.New("listener already running")
	}

	quitCh := make(chan struct{})
	msgCh := make(chan []byte, 10)
	l.quitch = quitCh
	l.msgch = msgCh
	l.running = true
	l.mu.Unlock()

	// Create a new listener based on the protocol
	ln, err := net.Listen(l.Protocol, l.Address+":"+l.Port)
	if err != nil {
		_ = l.UpdateStatus("Inactive", time.Now().Format("2006-01-02 15:04:05"))
		l.mu.Lock()
		l.running = false
		if l.quitch == quitCh {
			l.quitch = nil
		}
		if l.msgch == msgCh {
			l.msgch = nil
		}
		l.mu.Unlock()
		return err
	}

	l.mu.Lock()
	l.ln = ln
	l.mu.Unlock()

	_ = l.UpdateStatus("Active", time.Now().Format("2006-01-02 15:04:05"))

	go l.acceptLoop(ln, quitCh) // Start accepting connections in a separate goroutine

	<-quitCh // Wait for quit signal
	_ = ln.Close()
	close(msgCh)

	l.mu.Lock()
	if l.ln == ln {
		l.ln = nil
	}
	if l.quitch == quitCh {
		l.quitch = nil
	}
	if l.msgch == msgCh {
		l.msgch = nil
	}
	l.running = false
	l.mu.Unlock()

	return nil
}

func (l *Listener) Stop() {
	l.mu.Lock()
	if !l.running {
		l.mu.Unlock()
		return
	}

	quitCh := l.quitch
	ln := l.ln
	l.running = false
	l.mu.Unlock()

	if quitCh != nil {
		close(quitCh) // Signal to stop the listener
	}
	if ln != nil {
		_ = ln.Close() // Close the underlying network listener
	}

	l.UpdateStatus("Inactive", time.Now().Format("2006-01-02 15:04:05"))
}

func (l *Listener) ReinitializeRuntimeChannels() {
	l.quitch = make(chan struct{})
	l.msgch = make(chan []byte, 10)
}

func (l *Listener) UpdateStatus(status, lastCheckIn string) error {
	l.Status = status
	l.LastCheckIn = lastCheckIn
	l.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	db := Data.GetDB()
	if err := db.Save(l).Error; err != nil {
		log.Printf("Failed to update listener in database: %v", err)
		return err
	}

	return nil
}

func (l *Listener) GetMessages() <-chan []byte {
	return l.msgch
}

func (l *Listener) acceptLoop(ln net.Listener, quitCh <-chan struct{}) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-quitCh:
				return
			default:
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		l.UpdateStatus("Active", time.Now().Format("2006-01-02 15:04:05"))

		// Handle the connection in a new goroutine
		go l.handleConnection(conn)
	}
}

func (l *Listener) handleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	grunt := GetGruntByListenerAndAddress(l.ID, remoteAddr)
	id := ""

	if grunt == nil {
		id = uuid.New().String()
		Grunts[id] = NewGrunt(id, l.ID, remoteAddr, "Active", time.Now().Format("2006-01-02 15:04:05"))
	} else {
		id = grunt.ID
		if err := grunt.UpdateStatus("Active", time.Now().Format("2006-01-02 15:04:05")); err != nil {
			log.Printf("Failed to update grunt %s status: %v", id, err)
		}
	}

	l.UpdateStatus("Active", time.Now().Format("2006-01-02 15:04:05"))
	Listeners[l.ID] = l
	l.GruntCount = len(Grunts)

	gruntConnectionsMu.Lock()
	gruntConnections[id] = conn
	gruntConnectionsMu.Unlock()

	log.Printf("Accepted connection from %s (grunt: %s)", remoteAddr, id)

	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection closed by remote: %s", conn.RemoteAddr().String())
			} else {
				log.Printf("Error reading from connection: %v", err)
			}
			// Set grunt status to Inactive
			if Grunts[id] != nil {
				if err := Grunts[id].UpdateStatus("Inactive", time.Now().Format("2006-01-02 15:04:05")); err != nil {
					log.Printf("Failed to update grunt %s status: %v", id, err)
				}
			}
			if len(Grunts) == 0 {
				l.UpdateStatus("Active", time.Now().Format("2006-01-02 15:04:05"))
			}
			gruntConnectionsMu.Lock()
			delete(gruntConnections, id)
			gruntConnectionsMu.Unlock()
			return
		}

		incoming := string(buf[:n])

		if err := SaveGruntMessage(id, l.ID, incoming, false); err != nil {
			log.Printf("Failed to save message for grunt %s: %v", id, err)
		}

		select {
		case l.msgch <- buf[:n]:
		default:
			log.Printf("Dropping listener message publish for %s: channel full", l.ID)
		}
	}
}

func NewListener(id, address, port, protocol string) *Listener {
	db := Data.GetDB()
	defaultModel := models.DefaultModel{
		ID:        id,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	listener := &Listener{
		DefaultModel: defaultModel,
		Address:      address,
		Port:         port,
		Protocol:     protocol,
		Status:       "Active",
		LastCheckIn:  "",
		GruntCount:   0,
		quitch:       make(chan struct{}),
		msgch:        make(chan []byte, 10),
	}

	if err := db.Create(listener).Error; err != nil {
		log.Printf("Failed to save listener to database: %v", err)
	}

	return listener
}

func SendCommandToGrunt(gruntID, command string) error {
	gruntConnectionsMu.RLock()
	conn := gruntConnections[gruntID]
	gruntConnectionsMu.RUnlock()

	if conn == nil {
		return net.ErrClosed
	}

	_, err := conn.Write([]byte(command + "\n"))
	return err
}

func GetActiveListeners() []*Listener {
	active := make([]*Listener, 0, len(Listeners))
	for _, listener := range Listeners {
		if listener != nil && listener.Status == "Active" {
			active = append(active, listener)
		}
	}
	return active
}
