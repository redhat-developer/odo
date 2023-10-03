package db

import (
	"context"

	"google.golang.org/api/sheets/v4"
)

func SaveToSheet(ctx context.Context, sheetId string, data []interface{}) error {
	srv, err := getService(ctx)
	if err != nil {
		return err
	}
	readRange := "Template"
	valueRange := sheets.ValueRange{
		MajorDimension: "ROWS",
		Range:          "Template",
		Values: [][]interface{}{
			data,
		},
	}
	_, err = srv.Spreadsheets.Values.
		Append(sheetId, readRange, &valueRange).
		ValueInputOption("USER_ENTERED").Do()
	return err
}
