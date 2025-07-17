package vtec

import (
	"context"

	"github.com/metdatasystem/mds-awips/pkg/awips"
)

type Repository interface {
	GetEventByID(ctx context.Context, id int) (*VTECEvent, error)
	GetEventByVTEC(ctx context.Context, vtec awips.VTEC, year int) (*VTECEvent, error)
	CreateEvent(ctx context.Context, event *VTECEvent) error
}
