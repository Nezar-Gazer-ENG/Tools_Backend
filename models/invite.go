package models

type InviteInput struct {
	EventID string `json:"eventId"`
	Email   string `json:"email,omitempty"`
	Status  string `json:"status,omitempty"`
}
