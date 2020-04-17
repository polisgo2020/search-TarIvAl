package model

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/polisgo2020/search-tarival/index"
	"github.com/rs/zerolog/log"
)

// SaveDB save index to PostgreSQL database
func SaveDB(db *sql.DB, index index.ReverseIndex) {
	if _, err := db.Exec(`DELETE FROM positions`); err != nil {
		log.Error().Err(err).Msg("Execute delete table positions err")
	}
	if _, err := db.Exec(`DELETE FROM words`); err != nil {
		log.Error().Err(err).Msg("Execute delete table words err")
	}
	if _, err := db.Exec(`DELETE FROM files`); err != nil {
		log.Error().Err(err).Msg("Execute delete table files err")
	}

	var wID, fID int
	for word, data := range index {
		switch err := db.QueryRow(`SELECT w_id FROM words WHERE word=$1`, word).Scan(&wID); err {
		case sql.ErrNoRows:
			if _, err := db.Exec(`INSERT INTO words (word) VALUES ($1)`, word); err != nil {
				log.Error().Err(err).Msg("Execute insert word in table words err")
			}
			if err := db.QueryRow(`SELECT w_id FROM words WHERE word=$1`, word).Scan(&wID); err != nil {
				log.Error().Err(err).Msg("SELECT w_id FROM words err")
			}
		case nil:
		default:
			log.Error().Err(err).Msg("Err query w_id")
		}
		for _, file := range data {
			switch err := db.QueryRow(`SELECT f_id FROM files WHERE file=$1`, file.File).Scan(&fID); err {
			case sql.ErrNoRows:
				if _, err := db.Exec(`INSERT INTO files (file) VALUES ($1)`, file.File); err != nil {
					log.Error().Err(err).Msg("Execute insert file name in table files err")
				}
				if err := db.QueryRow(`SELECT f_id FROM files WHERE file=$1`, file.File).Scan(&fID); err != nil {
					log.Error().Err(err).Msg("SELECT f_id FROM files err")
				}
			case nil:
			default:
				log.Error().Err(err).Msg("Err query f_id")
			}
			for _, position := range file.Positions {
				if _, err := db.Exec(`INSERT INTO positions (w_id, f_id, position) VALUES ($1, $2, $3)`, wID, fID, position); err != nil {
					log.Error().Err(err).Int("wID", wID).Int("fID", fID).Int("pos", position).Msg("Execute insert position  in table positions err")
				}
			}
		}
	}
}

// LoadDB load index from PostgreSQL database
func LoadDB(db *sql.DB) index.ReverseIndex {
	ind := make(index.ReverseIndex)
	words := loadTwoColumnTable(db, "words")
	files := loadTwoColumnTable(db, "files")

	rows, err := db.Query("SELECT * FROM positions")
	if err != nil {
		log.Error().Err(err).Msg("SELECT * FROM positions err")
	}
	defer rows.Close()
	for rows.Next() {
		var wID, fID, position int
		err = rows.Scan(&wID, &fID, &position)
		if err != nil {
			log.Error().Err(err).Msg("Rows scan err")
		}

		if i := index.HasFileInIndex(ind[words[wID]], files[fID]); i != -1 {
			ind[words[wID]][i].Positions = append(ind[words[wID]][i].Positions, position)
		} else {
			item := index.WordIndex{
				File:      files[fID],
				Positions: []int{position},
			}
			ind[words[wID]] = append(ind[words[wID]], item)
		}
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msg("Rows err")
	}

	return ind
}

func loadTwoColumnTable(db *sql.DB, tableName string) map[int]string {
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		log.Error().Err(err).Msgf("SELECT * FROM %s err", tableName)
	}
	defer rows.Close()
	result := make(map[int]string)
	for rows.Next() {
		var id int
		var str string
		err = rows.Scan(&id, &str)
		if err != nil {
			log.Error().Err(err).Msg("Rows scan err")
		}
		result[id] = str
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msg("Rows err")
	}
	return result
}
