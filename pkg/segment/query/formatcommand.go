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

package query

import (
	"github.com/siglens/siglens/pkg/segment/formatter"
	"github.com/siglens/siglens/pkg/segment/structs"
	log "github.com/sirupsen/logrus"
)

// ProcessFormatCommand processes a format command in the query pipeline
// It converts the input records into a single record with a "search" field
// containing a formatted string representation according to the format options
func ProcessFormatCommand(letColReq *structs.LetColumnsRequest, records []map[string]interface{}, qid uint64) ([]map[string]interface{}, error) {
	if letColReq == nil || letColReq.FormatResults == nil {
		return records, nil
	}

	log.Infof("qid=%d, Executing format command", qid)

	// Set default values for format options if not specified
	formatReq := letColReq.FormatResults

	// Default value for MVSeparator
	if formatReq.MVSeparator == "" {
		formatReq.MVSeparator = "OR"
	}

	// Default value for EmptyString
	if formatReq.EmptyString == "" {
		formatReq.EmptyString = "NOT( )"
	}

	// Default values for row/column options
	if formatReq.RowColOptions == nil {
		formatReq.RowColOptions = &structs.RowColOptions{
			RowPrefix:       "(",
			ColumnPrefix:    "(",
			ColumnSeparator: "AND",
			ColumnEnd:       ")",
			RowSeparator:    "OR",
			RowEnd:          ")",
		}
	}

	// Call the formatter package to process the records
	formattedResults, err := formatter.FormatResults(records, formatReq, qid)
	if err != nil {
		return nil, err
	}

	// Important: completely replace all records with just our formatted result
	// This ensures we return only the single record with the search field
	return formattedResults, nil
}

// This is an example - adapt to your actual code structure
func ProcessLetColumnsRequest(letColReq *structs.LetColumnsRequest, records []map[string]interface{}, qid uint64) ([]map[string]interface{}, error) {
	if letColReq == nil {
		return records, nil
	}

	// Process format command - this must be handled specially as it replaces all results
	if letColReq.FormatResults != nil {
		// Format command completely replaces all results with a single record
		return ProcessFormatCommand(letColReq, records, qid)
	}

	// Process other let column commands only if format is not present
	// ... (process other commands)

	return records, nil
}
