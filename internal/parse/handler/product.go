package handler

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/metdatasystem/mds-awips/pkg/db"
)

type TextProduct db.TextProduct

type productHandler struct {
	Handler
}

func (handler *productHandler) Handle() (*TextProduct, error) {
	product := handler.awipsProduct

	id := fmt.Sprintf("%s-%s-%s-%s", product.Issued.UTC().Format("200601021504"), product.Office, product.WMO.Datatype, product.AWIPS.Original)

	if len(product.WMO.BBB) > 0 {
		id += "-" + product.WMO.BBB
	}

	textProduct := TextProduct{
		ProductID:  id,
		ReceivedAt: &handler.receivedAt,
		Issued:     &product.Issued,
		Source:     product.AWIPS.WFO,
		Data:       product.Text,
		WMO:        product.WMO.Datatype,
		AWIPS:      product.AWIPS.Original,
		BBB:        product.WMO.BBB,
	}

	rows, err := handler.db.Query(context.Background(), `
	INSERT INTO awips.products (product_id, received_at, issued, source, data, wmo, awips, bbb) VALUES
	($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at;
	`, id, handler.receivedAt, product.Issued, product.AWIPS.WFO, product.Text, product.WMO.Datatype, product.AWIPS.Original, product.WMO.BBB)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&textProduct.ID, &textProduct.CreatedAt)
		return &textProduct, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return nil, errors.New("no rows returned when creating new text product: " + rows.Err().Error())
}

func (product *TextProduct) isCorrection() bool {
	resent := regexp.MustCompile("...(RESENT|RETRANSMITTED|CORRECTED)")

	if len(resent.FindString(product.Data)) > 0 {
		return true
	}
	if len(product.BBB) > 0 && (string(product.BBB[0]) == "A" || string(product.BBB[0]) == "C") {
		return true
	}

	return false
}
