package models

type DefaultModel struct {
	ID        string `gorm:"primaryKey" json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Message struct {
	DefaultModel
	SenderID   		string `json:"sender_id"`
	ReceiverID 		string `json:"receiver_id"`
	IsServerMessage 	bool `json:"is_server_message"`
	Content   		string `json:"content"`
}