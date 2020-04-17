package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/polisgo2020/search-tarival/config"
	"github.com/polisgo2020/search-tarival/index"
	"github.com/polisgo2020/search-tarival/model"
	"github.com/polisgo2020/search-tarival/web"
	"github.com/urfave/cli/v2"
)

var cfg config.Config

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	cfg = config.Load()
	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("log level", cfg.LogLevel).
			Msg("")
	}
	zerolog.SetGlobalLevel(logLevel)

	app := &cli.App{
		Name:  "Searching and indexing",
		Usage: "make reverse index and search with it",
	}

	app.Commands = []*cli.Command{
		{
			Name:    "index",
			Aliases: []string{"i"},
			Usage:   "make a reverse index for directiory",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "path",
					Aliases:  []string{"p"},
					Required: true,
					Usage:    "path to directory",
				},
			},
			Subcommands: []*cli.Command{
				{
					Name:   "json",
					Usage:  "save index to json",
					Action: indexJSON,
				},
				{
					Name:   "db",
					Usage:  "save index to PostgeSQL darabase",
					Action: indexDB,
				},
			},
		},
		{
			Name:    "search",
			Aliases: []string{"s"},
			Usage:   "serching in directody with reverse index",
			Subcommands: []*cli.Command{
				{
					Name:   "json",
					Usage:  "load index from json",
					Action: searchJSON,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "index",
							Aliases:  []string{"i"},
							Required: true,
							Usage:    "path to reverse index",
						},
					},
				},
				{
					Name:   "db",
					Usage:  "load index from PostgeSQL darabase",
					Action: searchDB,
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("")
	}
}

func indexing(path string) index.ReverseIndex {
	if len(path) == 0 {
		log.Fatal().
			Err(errors.New("Path to folder not found")).
			Msg("")
	}
	index, err := index.IndexingFolder(path)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("")
	}
	return index
}

func indexJSON(c *cli.Context) error {
	index := indexing(c.String("path"))

	output, err := json.Marshal(index)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("")
	}
	if err := ioutil.WriteFile("index.json", output, 0666); err != nil {
		log.Fatal().
			Err(err).
			Msg("")
	}
	return nil
}

func indexDB(c *cli.Context) error {
	index := indexing(c.String("path"))

	db, err := sql.Open("postgres", cfg.PgSQL)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("")
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("")
	}

	model.SaveDB(db, index)
	return nil
}

func searchDB(c *cli.Context) error {

	db, err := sql.Open("postgres", cfg.PgSQL)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("")
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("")
	}

	Index := model.LoadDB(db)

	handleObjs := []web.HandleObject{
		web.HandleObject{
			Address:   "/",
			Tmp:       "web/templates/index.html",
			WithIndex: false,
		},
		web.HandleObject{
			Address:   "/result",
			Tmp:       "web/templates/result.html",
			WithIndex: true,
			Index:     Index,
		},
	}

	if err = web.ServerStart(cfg.Listen, 10*time.Second, handleObjs); err != nil {
		log.Fatal().
			Err(err).
			Msg("")
	}
	return nil
}

func searchJSON(c *cli.Context) error {

	indexName := c.String("index")

	Index, err := index.ReadIndexJSON(indexName)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("")
	}

	handleObjs := []web.HandleObject{
		web.HandleObject{
			Address:   "/",
			Tmp:       "web/templates/index.html",
			WithIndex: false,
		},
		web.HandleObject{
			Address:   "/result",
			Tmp:       "web/templates/result.html",
			WithIndex: true,
			Index:     Index,
		},
	}

	if err = web.ServerStart(cfg.Listen, 10*time.Second, handleObjs); err != nil {
		log.Fatal().
			Err(err).
			Msg("")
	}
	return nil
}
