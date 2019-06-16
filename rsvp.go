package main

// Rsvp contains all data for one RSVP
type Rsvp struct {
	ID         uint32 `json:"_id"`
	ShortCode  string `json:"shortcode"`
	Attending  bool   `json:"attending"`
	NumInvited int    `json:"numinvited"`
	MonConfirm int    `json:"monconfirm"`
	SunConfirm int    `json:"sunconfirm"`
}
