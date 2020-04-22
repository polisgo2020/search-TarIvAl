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
		fmt.Println(err, id, columnID, table, columnData, val)
		return 0, false, err
	}
}

// Insert - insert in table valsSlice values to columns
func Insert(db *sql.DB, table string, columns []string, valsSlice [][]string) error {
	var columnsStr string
	for i, column := range columns {
		if i != 0 {
			columnsStr += ", "
		}
		columnsStr += column
	}
	var inputValues string
	for j, vals := range valsSlice {
		if len(vals) != len(columns) {
			return errors.New("length columns and values not equal")
		}
		if j != 0 {
			inputValues += ", "
		}
		inputValues += "("
		for i, val := range vals {
			if i != 0 {
				inputValues += ", "
			}
			inputValues += val
		}
		inputValues += ")"
	}
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s`, table, columnsStr, inputValues)
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

// DownloadTable - download 2 column table into map
func DownloadTable(db *sql.DB, table string) (map[string]int, error) {
	query := fmt.Sprintf("SELECT * FROM %s", table)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]int)
	for rows.Next() {
		var id int
		var value string
		err = rows.Scan(&id, &value)
		if err != nil {
			return nil, err
		}
		result[value] = id
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}
