package model

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

// Delete - delete from table where value column = val
func Delete(db *sql.DB, table, column, val string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s = %s`, table, column, val)
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

// CheckAndInsert - check val in table columnData, insert val in columnData if columnData havn't val.
// Return val id, had val in columnData before func and err if it has
func CheckAndInsert(db *sql.DB, table, columnData, columnID, val string) (int, bool, error) {
	var id int
	var err error
	querySelect := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = '%s'`, columnID, table, columnData, val)
	switch err = db.QueryRow(querySelect).Scan(&id); err {
	case sql.ErrNoRows:
		queryInsert := fmt.Sprintf(`INSERT INTO %s (%s) VALUES ('%s')`, table, columnData, val)
		_, err = db.Exec(queryInsert)
		if err != nil {
			return 0, false, err
		}
		err = db.QueryRow(querySelect).Scan(&id)
		if err != nil {
			return 0, false, err
		}
		return id, false, nil
	case nil:
		return id, true, nil
	default:
		return 0, false, err
	}
}

// InsertRow - insert 1 row in table and columns will = vals
func InsertRow(db *sql.DB, table string, columns, vals []string) error {
	if len(vals) != len(columns) {
		return errors.New("length columns and values not equal")
	}
	var columnsStr string
	for i, column := range columns {
		if i != 0 {
			columnsStr += ", "
		}
		columnsStr += column
	}
	var valsStr string
	for i, val := range vals {
		if i != 0 {
			valsStr += ", "
		}
		valsStr += val
	}
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, table, columnsStr, valsStr)
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
