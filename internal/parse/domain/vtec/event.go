package vtec

import (
	"time"

	"github.com/twpayne/go-geos"
)

// VTEC Event
type VTECEvent struct {
	ID           int       `json:"id,omitempty"`
	EventID      string    `json:"event_id"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
	Issued       time.Time `json:"issued"`
	Starts       time.Time `json:"starts,omitempty"`
	Expires      time.Time `json:"expires"`
	Ends         time.Time `json:"ends,omitempty"`
	EndInitial   time.Time `json:"end_initial,omitempty"`
	Class        string    `json:"class"`
	Phenomena    string    `json:"phenomena"`
	WFO          string    `json:"wfo"`
	Significance string    `json:"significance"`
	EventNumber  int       `json:"event_number"`
	Year         int       `json:"year"`
	Title        string    `json:"title"`
	IsEmergency  bool      `json:"is_emergency"`
	IsPDS        bool      `json:"is_pds"`
	PolygonStart *geos.Geom
}
