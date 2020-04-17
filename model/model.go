package model

import (
	"database/sql"

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
