package sql

import (
	"database/sql/driver"
	"net/http"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

var testCasesGetCollectionAsOptionsFilterable = []testCase{
	{
		Kind: "success",
		Name: "it succeeds when equipment table contains more than one column",
		DescriptorFields: []string{
			commonEquipmentDescriptorFields,
			commonMaintenanceDescriptorFields,
		},
		TableSchema: commonEquipmentTableSchema,
		ColumnNames: []string{
			"equipment\x00id",
			"equipment\x00name",
		},
		RowsAsCsv: "1,Stainless Steel Mash Tun (50L)",
		ExpectedResults: `{
  "id": "1",
  "name": "Stainless Steel Mash Tun (50L)"
}`,
		ExpectedQueries: func(mock sqlmock.Sqlmock, columns []string, rowsAsCsv string, args ...driver.Value) {
			rows := sqlmock.NewRows(columns).
				FromCSVString(rowsAsCsv)
			mock.ExpectQuery("SELECT (.+), (.+) FROM (.+) WHERE (.+) LIKE (.+)").
				WillReturnRows(rows)
		},
		Request: func() *http.Request {
			req, _ := http.NewRequest("GET", "/equipment/options?filter=stain", nil)
			return req
		},
	},
}
