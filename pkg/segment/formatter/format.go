// Copyright (c) 2021-2024 SigScalr, Inc.
//
// This file is part of SigLens Observability Solution
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package formatter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/siglens/siglens/pkg/segment/structs"
	log "github.com/sirupsen/logrus"
)

// FormatResults processes a FormatResultsRequest and transforms the given records
// by adding a "search" field containing a formatted string representation
// while preserving the original fields
func FormatResults(records []map[string]interface{}, formatReq *structs.FormatResultsRequest, qid uint64) ([]map[string]interface{}, error) {
	if formatReq == nil {
		return nil, fmt.Errorf("format request is nil")
	}

	log.Infof("qid=%d, Processing format command with %d records", qid, len(records))

	if len(records) == 0 {
		// If no records, create a new record with just the empty string
		result := make(map[string]interface{})
		result["search"] = formatReq.EmptyString
		return []map[string]interface{}{result}, nil
	}

	// Limit the number of records if maxResults is specified
	maxRecords := len(records)
	if formatReq.MaxResults > 0 && uint64(maxRecords) > formatReq.MaxResults {
		maxRecords = int(formatReq.MaxResults)
		log.Debugf("qid=%d, Format command limiting from %d to %d records", qid, len(records), maxRecords)
	}

	rowColOpts := formatReq.RowColOptions

	// Format all records
	rowExpressions := make([]string, 0, maxRecords)
	for i := 0; i < maxRecords; i++ {
		record := records[i]

		// Skip records with no fields
		if len(record) == 0 {
			continue
		}

		// Get sorted keys for consistent output
		keys := make([]string, 0, len(record))
		for key := range record {
			// Skip internal fields
			if key == "_raw" || key == "search" {
				continue
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)

		// Build column expressions for this record
		colExpressions := make([]string, 0, len(keys))
		for _, key := range keys {
			// Format the value based on its type
			formattedValue := formatValue(record[key], formatReq.MVSeparator)
			colExpr := fmt.Sprintf("%s=\"%s\"", key, formattedValue)
			colExpressions = append(colExpressions, colExpr)
		}

		// Combine column expressions for this record
		if len(colExpressions) > 0 {
			rowExpr := rowColOpts.ColumnPrefix
			rowExpr += strings.Join(colExpressions, " "+rowColOpts.ColumnSeparator+" ")
			rowExpr += rowColOpts.ColumnEnd
			rowExpressions = append(rowExpressions, rowExpr)
		}
	}

	// Combine all row expressions
	var formattedStr string
	if len(rowExpressions) > 0 {
		formattedStr = rowColOpts.RowPrefix
		formattedStr += strings.Join(rowExpressions, " "+rowColOpts.RowSeparator+" ")
		formattedStr += rowColOpts.RowEnd
	} else {
		formattedStr = formatReq.EmptyString
	}

	// Create new result records with original data + search field
	result := make([]map[string]interface{}, 0, maxRecords)
	for i := 0; i < maxRecords && i < len(records); i++ {
		// Create a flat structure with search field first
		newRecord := make(map[string]interface{})

		// Add the search field first
		newRecord["search"] = formattedStr

		// Add all original fields
		for k, v := range records[i] {
			if k != "search" && k != "_display_order" {
				newRecord[k] = v
			}
		}

		// Add display order metadata
		displayOrder := []string{"search"}
		for k := range records[i] {
			if k != "search" && k != "_display_order" {
				displayOrder = append(displayOrder, k)
			}
		}
		newRecord["_display_order"] = displayOrder

		result = append(result, newRecord)
	}

	// Before returning the results
	log.Infof("Format command result with search field: %v", result[0]["search"])
	return result, nil
}

// formatValue formats a value based on its type
func formatValue(value interface{}, mvSeparator string) string {
	if value == nil {
		return "null"
	}

	// If it's a slice, join the values with the specified separator
	switch v := value.(type) {
	case []interface{}:
		strSlice := make([]string, len(v))
		for i, item := range v {
			strSlice[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(strSlice, " "+mvSeparator+" ")
	case []string:
		return strings.Join(v, " "+mvSeparator+" ")
	}

	// For other types, convert to string
	return fmt.Sprintf("%v", value)
}
