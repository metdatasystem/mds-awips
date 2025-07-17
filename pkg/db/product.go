package db

import "time"

// Text Product
type TextProduct struct {
	ID         int        `json:"id"`
	ProductID  string     `json:"product_id"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	ReceivedAt *time.Time `json:"received_at"`
	Issued     *time.Time `json:"issued"`
	Source     string     `json:"source"`
	Data       string     `json:"data"`
	WMO        string     `json:"wmo"`
	AWIPS      string     `json:"awips"`
	BBB        string     `json:"bbb"`
}
