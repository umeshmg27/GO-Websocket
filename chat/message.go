package chat

type MessageType int

const (
	MessageTypeText MessageType = iota
	MessageTypePrivate
	MessageTypeJoin
	MessageTypeLeave
	MessageTypeError
)

type Message struct {
	Type        MessageType `json:"type"`
	Text        string      `json:"text"`
	SenderID    string      `json:"sender_id"`
	SenderName  string      `json:"sender_name"`
	RecipientID string      `json:"recipient_id,omitempty"`
}
