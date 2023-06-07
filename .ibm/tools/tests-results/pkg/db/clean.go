package db

import (
	"context"
	"fmt"

	"google.golang.org/api/sheets/v4"
)

func Clean(ctx context.Context, sheetId string, beforeDate string) error {
	srv, err := getService(ctx)
	if err != nil {
		return err
	}
	range_ := "Template!A2:A"

	result, err := srv.Spreadsheets.Values.Get(sheetId, range_).Do()
	if err != nil {
		return err
	}
	var rowNumber int64 = 2
	for _, row := range result.Values {
		value := row[0].(string)
		if value >= beforeDate {
			if rowNumber > 2 {
				fmt.Printf("Delete rows from 2 to %d\n", rowNumber-1)
			}
			break
		}
		rowNumber++
	}

	if rowNumber < 3 {
		return nil
	}

	var requests []*sheets.Request
	requests = append(requests, &sheets.Request{
		DeleteDimension: &sheets.DeleteDimensionRequest{
			Range: &sheets.DimensionRange{
				Dimension:  "ROWS",
				StartIndex: 1,             // 0-index, inclusive
				EndIndex:   rowNumber - 1, // 0-index, exclusive
			},
		},
	})
	req := sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}
	_, err = srv.Spreadsheets.BatchUpdate(sheetId, &req).Do()
	return err
}
