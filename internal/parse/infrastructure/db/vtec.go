package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metdatasystem/mds-awips/internal/parse/domain/vtec"
	"github.com/metdatasystem/mds-awips/pkg/awips"
)

type vtecRepository struct {
	db *pgxpool.Pool
}

func NewVTECRepository(db *pgxpool.Pool) *vtecRepository {
	return &vtecRepository{db: db}
}

// Get a VTEC event by its ID.
func (r *vtecRepository) GetEventByID(ctx context.Context, id int) (*vtec.VTECEvent, error) {
	rows, err := r.db.Query(ctx, `SELECT * FROM vtec_events WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // No event found with the given ID
	}
	var event vtec.VTECEvent
	err = rows.Scan(
		&event.ID,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.Issued,
		&event.Starts,
		&event.Expires,
		&event.Ends,
		&event.EndInitial,
		&event.Class,
		&event.Phenomena,
		&event.WFO,
		&event.Significance,
		&event.EventNumber,
		&event.Year,
		&event.Title,
		&event.IsEmergency,
		&event.IsPDS,
		&event.PolygonStart,
	)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// Get a VTEC event by its VTEC properties and a year.
func (r *vtecRepository) GetEventByVTEC(ctx context.Context, v awips.VTEC, year int) (*vtec.VTECEvent, error) {
	rows, err := r.db.Query(ctx, `
			SELECT * FROM vtec.events WHERE
			wfo = $1 AND phenomena = $2 AND significance = $3 AND event_number = $4 AND year = $5
			`, v.WFO, v.Phenomena, v.Significance, v.EventNumber, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // No event found with the given ID
	}
	var event vtec.VTECEvent
	err = rows.Scan(
		&event.ID,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.Issued,
		&event.Starts,
		&event.Expires,
		&event.Ends,
		&event.EndInitial,
		&event.Class,
		&event.Phenomena,
		&event.WFO,
		&event.Significance,
		&event.EventNumber,
		&event.Year,
		&event.Title,
		&event.IsEmergency,
		&event.IsPDS,
		&event.PolygonStart,
	)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// Inserts an event into the database.
func (r *vtecRepository) CreateEvent(ctx context.Context, event *vtec.VTECEvent) error {
	_, err := r.db.Exec(ctx, `
	INSERT INTO vtec.events(issued, starts, expires, ends, end_initial, class, phenomena, wfo, 
	significance, event_number, year, title, is_emergency, is_pds, polygon_start) VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);
	`, event.Issued, event.Starts, event.Expires, event.Ends, event.EndInitial, event.Class,
		event.Phenomena, event.WFO, event.Significance, event.EventNumber, event.Year, event.Title,
		event.IsEmergency, event.IsPDS, event.PolygonStart)
	return err
}
