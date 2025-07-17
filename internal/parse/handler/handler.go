package handler

import (
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metdatasystem/mds-awips/internal/parse/infrastructure/db"
	"github.com/metdatasystem/mds-awips/pkg/awips"
	"github.com/metdatasystem/mds-awips/pkg/logger"
)

var (
	vtecRoute = regexp.MustCompile("(MWW|FWW|CFW|TCV|RFW|FFA|SVR|TOR|SVS|SMW|MWS|NPW|WCN|WSW|EWW|FLS)")
	mcdRoute  = regexp.MustCompile("(SWOMCD)")
)

var routes = []Route{
	// VTEC Products
	{
		Name:    "VTEC Handler",
		Match:   func(product *awips.TextProduct) bool { return vtecRoute.MatchString(product.AWIPS.Product) },
		Handler: func(handler Handler) HandlerFunc { return &vtecHandler{handler, db.NewVTECRepository(handler.db)} },
	},
}

type Route struct {
	Name    string
	Match   func(product *awips.TextProduct) bool
	Handler func(handler Handler) HandlerFunc
}

type Handler struct {
	db           *pgxpool.Pool
	log          *logger.Logger
	text         string
	receivedAt   time.Time
	product      TextProduct
	awipsProduct *awips.TextProduct
}

type HandlerFunc interface {
	Handle()
	// Commit() error
}

func New(db *pgxpool.Pool, minlog int, text string, receivedAt time.Time) *Handler {
	log := logger.New(db, slog.Level(minlog))

	return &Handler{
		db:         db,
		log:        &log,
		text:       text,
		receivedAt: receivedAt,
	}
}

func (handler *Handler) Handle() {
	log := handler.log
	text := handler.text

	// Get the WMO header
	wmo, err := awips.ParseWMO(text)
	if err != nil {
		log.Error(err.Error())
		return
	}

	// Get the AWIPS header
	awipsHeader, err := awips.ParseAWIPS(text)
	if err != nil {
		log.Debug(err.Error())
	}

	// No point continuing if there is no AWIPS header
	if awipsHeader.Original == "" {
		log.Info("AWIPS header not found. Product will not be stored.")
		return
	} else {
		log.With("awips", awipsHeader.Original)
	}

	// Find the issue time
	issued, err := awips.GetIssuedTime(text)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if issued.IsZero() {
		log.Info("Product does not contain issue date. Defaulting to now (UTC)")
		issued = time.Now().UTC()
	}

	// Segment the product
	splits := strings.Split(text, "$$")

	segments := []awips.TextProductSegment{}

	for _, segment := range splits {
		segment = strings.TrimSpace(segment)

		// Assume the segment is the end of the product if it is shorter than 10 characters
		if len(segment) < 20 {
			continue
		}

		ugc, err := awips.ParseUGC(segment)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		expires := time.Now().UTC()
		if ugc != nil {
			expires = time.Date(issued.Year(), issued.Month(), ugc.Expires.Day(), ugc.Expires.Hour(), ugc.Expires.Minute(), 0, 0, time.UTC)
			if ugc.Expires.Day() > wmo.Issued.Day() && ugc.Expires.Day() == 1 {
				expires = expires.AddDate(0, 1, 0)
			}
			ugc.Merge(issued)
		}

		// Find any VTECs that the segment may have
		vtec, e := awips.ParseVTEC(segment)
		if len(e) != 0 {
			for _, er := range e {
				log.Error(er.Error())
			}
			continue
		}

		latlon, err := awips.ParseLatLon(segment)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		tags, e := awips.ParseTags(segment)
		if len(e) != 0 {
			for _, er := range e {
				log.Error(er.Error())
			}
		}

		tml, err := awips.ParseTML(segment, issued)
		if err != nil {
			log.Warn("failed to parse TML: " + err.Error())
		}

		segments = append(segments, awips.TextProductSegment{
			Text:    segment,
			VTEC:    vtec,
			UGC:     ugc,
			Expires: expires,
			LatLon:  latlon,
			Tags:    tags,
			TML:     tml,
		})

	}

	handler.awipsProduct = &awips.TextProduct{
		Text:     handler.text,
		WMO:      wmo,
		AWIPS:    awipsHeader,
		Issued:   issued,
		Office:   wmo.Office,
		Product:  awipsHeader.Product,
		Segments: segments,
	}

	committedProduct := false

	// Process the product matching it to any routes
	for _, route := range routes {
		if route.Match(handler.awipsProduct) {
			if !committedProduct {
				pHandler := productHandler{*handler}
				product, err := pHandler.Handle()
				if err != nil {
					log.Error("failed to handle product", "error", err)
					continue
				}
				handler.product = *product
			}
			h := route.Handler(*handler)
			h.Handle()
			// if err != nil {
			// 	handler.log.Error("failed to save log", "error", err, "handler", route.Name)
			// }
		}
	}
}

func (handler *Handler) SaveLog() error {
	return handler.log.Commit()
}
