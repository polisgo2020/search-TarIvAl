package model

import (
	"fmt"

	"github.com/go-pg/pg/v9"
	_ "github.com/lib/pq"
)

type Word struct {
	Id   int    `pg:"w_id,pk"`
	Word string `pg:"word"`
}

type File struct {
	Id   int    `pg:"f_id,pk"`
	File string `pg:"name_file"`
}

type Position struct {
	Wid      int    `pg:"w_id,pk"`
	Fid      int    `pg:"f_id,pk"`
	Position int `pg:"position"`
}

// Delete - delete from table where value column = val
func Delete(db *pg.DB, table, column, val string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s = ?`, table, column)
	_, err := db.Exec(query, val)
	if err != nil {
		return err
	}
	return nil
}

func (w *Word) CheckAndInsert(db *pg.DB) (bool, error) {
	ok, err := db.Model(w).
		Where("word = ?", w.Word).
		SelectOrInsert()
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func (f *File) CheckAndInsert(db *pg.DB) (bool, error) {
	ok, err := db.Model(f).
		Where("name_file = ?", f.File).
		SelectOrInsert()
	if err != nil {
		return ok, err
	}
	return ok, nil
}

// Insert - insert in table valsSlice values to columns
func Insert(db *pg.DB, buffer []Position) error {
	err := db.Insert(&buffer)
	if err != nil {
		return err
	}
	return nil
}

func SelectWords(db *pg.DB) (map[string]int, error) {
	result := make(map[string]int)
	var words []Word
	err := db.Model(&words).Select()
	if err != nil {
		return nil, err
	}
	for _, word := range words {
		result[word.Word] = word.Id
	}
	return result, nil
}

func SelectFiles(db *pg.DB) (map[int]string, error) {
	result := make(map[int]string)
	var files []File
	err := db.Model(&files).Select()
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		result[file.Id] = file.File
	}
	return result, nil
}

func (w *Word) SelectRow(db *pg.DB) error {
	return db.Model(w).Where("word = ?", w.Word).Select()
}

func SelectPositions(db *pg.DB, w_id int) ([]Position, error) {
	var positions []Position
	if err := db.Model(&positions).Where("w_id = ?", w_id).Select(); err != nil {
		return nil, err
	}
	return positions, nil
}
