package db

import (
	"time"

	"github.com/twpayne/go-geos"
)

// VTEC UGC Relation
type VTECUGC struct {
	ID           int       `json:"id"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
	WFO          string    `json:"wfo"`
	Phenomena    string    `json:"phenomena"`
	Significance string    `json:"significance"`
	EventNumber  int       `json:"event_number"`
	UGC          int       `json:"ugc"`
	Issued       time.Time `json:"issued"`
	Starts       time.Time `json:"starts,omitempty"`
	Expires      time.Time `json:"expires"`
	Ends         time.Time `json:"ends,omitempty"`
	EndInitial   time.Time `json:"end_initial,omitempty"`
	Action       string    `json:"action"`
	Latest       int       `json:"latest"`
	Year         int       `json:"year"`
}

// VTEC Event Update
type VTECUpdate struct {
	ID            int        `json:"id"`
	CreatedAt     time.Time  `json:"created_at,omitempty"`
	Issued        time.Time  `json:"issued"`
	Starts        time.Time  `json:"starts,omitempty"`
	Expires       time.Time  `json:"expires"`
	Ends          time.Time  `json:"ends,omitempty"`
	Text          string     `json:"text"`
	Product       string     `json:"product"`
	WFO           string     `json:"wfo"`
	Action        string     `json:"action"`
	Class         string     `json:"class"`
	Phenomena     string     `json:"phenomena"`
	Significance  string     `json:"significance"`
	EventNumber   int        `json:"event_number"`
	Year          int        `json:"year"`
	Title         string     `json:"title"`
	IsEmergency   bool       `json:"is_emergency"`
	IsPDS         bool       `json:"is_pds"`
	Polygon       *geos.Geom `json:"polygon,omitempty"`
	Direction     *int       `json:"direction"`
	Location      *geos.Geom `json:"location"`
	Speed         *int       `json:"speed"`
	SpeedText     *string    `json:"speed_text"`
	TMLTime       *time.Time `json:"tml_time"`
	UGC           []string   `json:"ugc"`
	Tornado       string
	Damage        string
	HailThreat    string
	HailTag       string
	WindThreat    string
	WindTag       string
	FlashFlood    string
	RainfallTag   string
	FloodTagDam   string
	SpoutTag      string
	SnowSquall    string
	SnowSquallTag string
}
