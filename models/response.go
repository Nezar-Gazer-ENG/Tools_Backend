package models

type ResponseInput struct {
	EventID string `json:"eventId"`
	Email   string `json:"email"`
	Status  string `json:"status"`
}
