package model

type RunJunitMessage struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

func NewRunJunitMessage(message string) *RunJunitMessage {
	return &RunJunitMessage{Message: message, Text: message}
}
