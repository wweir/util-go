package influx

import (
	"fmt"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/wweir/utils/util"
)

type Record struct {
	Tag map[string]string
	Row map[string]interface{}
}

// RespToRecords turn responce into records
func RespToRecords(resp *client.Response, e error) ([]*Record, error) {
	if err := util.FirstErr(e, resp); err != nil {
		return nil, err
	} else if count := len(resp.Results); count != 1 {
		return nil, fmt.Errorf("influx return results should be 1 but not %d", count)
	} else if count := len(resp.Results[0].Series); count == 0 {
		return nil, fmt.Errorf("no record found")
	} else if len(resp.Results[0].Series[0].Values) != 1 {
		return nil, fmt.Errorf("parse multi line record fail")
	}

	records := []*Record{}
	for _, serie := range resp.Results[0].Series {
		keys := serie.Columns
		vals := serie.Values
		record := &Record{Tag: serie.Tags, Row: map[string]interface{}{}}
		for fieldIdx := range keys {
			record.Row[keys[fieldIdx]] = vals[0][fieldIdx]
		}
		records = append(records, record)
	}

	return records, nil
}
