package bonfirec2

import (
	"log"
	"time"

	"github.com/Jojojojodr/bonfirec2/models"
	"github.com/google/uuid"
)

var Grunts = make(map[string]*Grunt)

type Grunt struct {
	models.DefaultModel
	ListenerID  string `json:"listener_id"`
	Address     string `json:"address"`
	Status      string `json:"status"`
	LastCheckIn string `json:"last_check_in"`
}

func (g *Grunt) UpdateStatus(status, lastCheckIn string) error {
	g.Status = status
	g.LastCheckIn = lastCheckIn
	g.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	db := Data.GetDB()
	if err := db.Save(g).Error; err != nil {
		log.Printf("Failed to update grunt in database: %v", err)
		return err
	}

	return nil
}

func (g *Grunt) Delete() error {
	db := Data.GetDB()
	if err := db.Delete(g).Error; err != nil {
		log.Printf("Failed to delete grunt from database: %v", err)
		return err
	}
	return nil
}

func (g *Grunt) GetMessages() ([]*models.Message, error) {
	db := Data.GetDB()
	var messages []*models.Message
	if err := db.Where("sender_id = ? OR receiver_id = ?", g.ID, g.ID).Order("created_at ASC").Find(&messages).Error; err != nil {
		log.Printf("Failed to retrieve messages from database: %v", err)
		return nil, err
	}
	return messages, nil
}

func NewGrunt(id, listenerID, address, status, lastCheckIn string) *Grunt {
	db := Data.GetDB()
	defaultModel := models.DefaultModel{
		ID:        id,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	grunt := &Grunt{
		DefaultModel: defaultModel,
		ListenerID:   listenerID,
		Address:      address,
		Status:       status,
		LastCheckIn:  lastCheckIn,
	}

	if err := db.Create(grunt).Error; err != nil {
		log.Printf("Failed to save grunt to database: %v", err)
	}
	Grunts[id] = grunt

	return grunt
}

func SaveGruntMessage(senderID, receiverID, content string, isServerMessage bool) error {
	db := Data.GetDB()
	msgID := uuid.New().String()
	msg := &models.Message{
		DefaultModel: models.DefaultModel{
			ID:        msgID,
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
		},
		SenderID:        senderID,
		ReceiverID:      receiverID,
		IsServerMessage: isServerMessage,
		Content:         content,
	}

	if err := db.Create(msg).Error; err != nil {
		return err
	}
	return nil
}

func GetGruntByID(id string) (*Grunt, error) {
	db := Data.GetDB()
	var grunt Grunt
	if err := db.First(&grunt, "id = ?", id).Error; err != nil {
		log.Printf("Failed to retrieve grunt from database: %v", err)
		return nil, err
	}
	return &grunt, nil
}

func GetActiveGrunts() []*Grunt {
	active := make([]*Grunt, 0, len(Grunts))
	for _, grunt := range Grunts {
		if grunt != nil && grunt.Status == "Active" {
			active = append(active, grunt)
		}
	}
	return active
}

func SetAllGruntsInactive() {
	lastCheckIn := time.Now().Format("2006-01-02 15:04:05")
	db := Data.GetDB()
	if err := db.Model(&Grunt{}).
		Where("status <> ?", "Inactive").
		Updates(map[string]any{
			"status":        "Inactive",
			"last_check_in": lastCheckIn,
			"updated_at":    lastCheckIn,
		}).Error; err != nil {
		log.Printf("Failed to set grunts inactive during shutdown: %v", err)
	}

	for _, grunt := range Grunts {
		if grunt == nil {
			continue
		}
		grunt.Status = "Inactive"
		grunt.LastCheckIn = lastCheckIn
		grunt.UpdatedAt = lastCheckIn
	}
}

func GetLatestMessages() []*models.Message {
	db := Data.GetDB()
	var messages []*models.Message
	if err := db.Order("created_at DESC").Limit(10).Find(&messages).Error; err != nil {
		log.Printf("Failed to retrieve latest messages from database: %v", err)
		return nil
	}
	return messages
}

func GetAllGrunts() ([]*Grunt, error) {
	db := Data.GetDB()
	var grunts []*Grunt
	if err := db.Order("created_at DESC").Find(&grunts).Error; err != nil {
		log.Printf("Failed to retrieve grunts from database: %v", err)
		return nil, err
	}
	return grunts, nil
}

func GetGruntsByListenerID(listenerID string) ([]*Grunt, error) {
	db := Data.GetDB()
	var grunts []*Grunt
	if err := db.Where("listener_id = ?", listenerID).Order("created_at DESC").Find(&grunts).Error; err != nil {
		log.Printf("Failed to retrieve grunts for listener %s: %v", listenerID, err)
		return nil, err
	}
	return grunts, nil
}

func GetGruntByListenerAndAddress(listenerID, address string) *Grunt {
	for _, grunt := range Grunts {
		if grunt == nil {
			continue
		}
		if grunt.ListenerID == listenerID && grunt.Address == address {
			return grunt
		}
	}
	return nil
}
