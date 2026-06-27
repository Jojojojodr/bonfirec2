package bonfirec2

import (
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
}

func (l *Listener) Start() error {
	if l.quitch == nil {
		l.quitch = make(chan struct{})
	}
	if l.msgch == nil {
		l.msgch = make(chan []byte, 10)
	}

	// Create a new listener based on the protocol
	ln, err := net.Listen(l.Protocol, l.Address+":"+l.Port)
	if err != nil {
		return err
	}

	defer ln.Close()
	l.ln = ln

	go l.acceptLoop() // Start accepting connections in a separate goroutine

	<-l.quitch     // Wait for quit signal
	close(l.msgch) // Close the message channel when done

	return nil
}

func (l *Listener) Stop() {
	close(l.quitch) // Signal to stop the listener
	delete(Listeners, l.ID)
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

func (l *Listener) acceptLoop() {
	for {
		conn, err := l.ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		l.Status = "Active"
		l.LastCheckIn = time.Now().Format("2006-01-02 15:04:05")

		// Handle the connection in a new goroutine
		go l.handleConnection(conn)
	}
}

func (l *Listener) handleConnection(conn net.Conn) {
	defer conn.Close()

	id := uuid.New().String()

	if Grunts[id] == nil {
		Grunts[id] = NewGrunt(id, l.ID, conn.RemoteAddr().String(), "Active", time.Now().Format("2006-01-02 15:04:05"))
		l.UpdateStatus("Active", time.Now().Format("2006-01-02 15:04:05"))
		Listeners[l.ID] = l
		l.GruntCount = len(Grunts)
	}

	gruntConnectionsMu.Lock()
	gruntConnections[id] = conn
	gruntConnectionsMu.Unlock()

	log.Printf("Accepted connection from %s", conn.RemoteAddr().String())

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
				l.UpdateStatus("Inactive", time.Now().Format("2006-01-02 15:04:05"))
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
		Status:       "Inactive",
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
