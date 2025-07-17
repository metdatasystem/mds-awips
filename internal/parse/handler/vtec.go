package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/metdatasystem/mds-awips/internal/parse/domain/vtec"
	"github.com/metdatasystem/mds-awips/internal/parse/util"
	"github.com/metdatasystem/mds-awips/pkg/awips"
	"github.com/twpayne/go-geos"
)

type vtecHandler struct {
	Handler
	repo vtec.Repository
}

func (handler *vtecHandler) Handle() {
	awipsProduct := handler.awipsProduct
	product := handler.product
	log := handler.log

	// Set the text product's ID in the logger
	log.With("product", product.ProductID)

	// Process each segment separately since they reference different UGC areas
	for i, segment := range awipsProduct.Segments {

		// This segment does not have a VTEC so we can skip it
		if len(segment.VTEC) == 0 {
			log.Info(fmt.Sprintf("Product segment %d does not have VTECs. Skipping...", i))
			continue
		}

		// Go through each VTEC in the segment and process it
		for _, vtec := range segment.VTEC {
			// Skip test and routine products
			if vtec.Class == "T" || vtec.Action == "ROU" {
				continue
			}

			// Find or create the VTEC event
			event, err := handler.getOrCreateVTECEvent(vtec, product, segment)
			if err != nil {
				log.Error("failed to get or create VTEC event", "error", err, "vtec", vtec.Original)
				continue
			}

			handler.updateTimes(vtec, event, segment)
		}
	}
}

func (handler *vtecHandler) SaveLog() error {

	return nil
}

// Try and find the VTEC event in the database or create a new one if it doesn't exist.
func (handler *vtecHandler) getOrCreateVTECEvent(v awips.VTEC, product TextProduct, segment awips.TextProductSegment) (*vtec.VTECEvent, error) {
	var year int
	if v.Start != nil {
		year = v.Start.Year()
	} else {
		year = product.Issued.Year()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	event, err := handler.repo.GetEventByVTEC(ctx, v, year)
	if err != nil {
		return nil, err
	}

	if event != nil {
		// The database needs a start and end time but VTECs may not have one.
		if v.Start == nil {
			// Use the issue time for the start time
			v.Start = product.Issued
		}
		if v.End == nil {
			// Use the expiry of the product for the end time
			v.End = &segment.Expires
		}

		// Create the polygon if there is one.
		var polygon *geos.Geom
		if segment.LatLon != nil {
			p := util.PolygonFromAwips(*segment.LatLon.Polygon)
			polygon = p
		}

		// Build the event
		event = &vtec.VTECEvent{
			Issued:       *product.Issued,
			Starts:       *v.Start,
			Expires:      segment.UGC.Expires,
			Ends:         *v.End,
			EndInitial:   *v.End,
			Class:        v.Class,
			Phenomena:    v.Phenomena,
			WFO:          v.WFO,
			Significance: v.Significance,
			EventNumber:  v.EventNumber,
			Year:         year,
			Title:        v.Title(segment.IsEmergency()),
			IsEmergency:  segment.IsEmergency(),
			IsPDS:        segment.IsPDS(),
			PolygonStart: polygon,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Create the event in the database
		err = handler.repo.CreateEvent(ctx, event)
	}

	err := 

	return event, err
}

// Update the times of the event relative to the action
func (handler *vtecHandler) updateTimes(vtec awips.VTEC, event *vtec.VTECEvent, segment awips.TextProductSegment) {
	product := handler.product

	// The product expires at the UGC expiry time
	var end time.Time
	if vtec.End == nil {
		end = segment.UGC.Expires
		handler.log.Info("VTEC end time is nil. Defaulting to UGC expiry time.")
	} else {
		end = *vtec.End
	}

	switch vtec.Action {
	case "CAN":
		fallthrough
	case "UPG":
		event.Expires = segment.UGC.Expires
		event.Ends = product.Issued.UTC()
	case "EXP":
		event.Expires = end
		event.Ends = end
	case "EXT":
		fallthrough
	case "EXB":
		event.Ends = end
		event.Expires = segment.UGC.Expires
	default:
		// NEW and CON
		if event.Ends.Before(end) {
			event.Ends = end
		}
		if event.Expires.Before(segment.Expires) {
			event.Expires = segment.Expires
		}
	}
}
