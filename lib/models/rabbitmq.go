package models

// Server side rabbitmq message with the direction payload
type RmqMessage struct {
	Swiper    string `json:"swiper"`
	Swipee    string `json:"swipee"`
	Comment   string `json:"comment"`
	Direction string `json:"direction"`
}
