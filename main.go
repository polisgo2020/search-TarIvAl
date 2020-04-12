package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/polisgo2020/search-tarival/config"
	"github.com/polisgo2020/search-tarival/index"
	"github.com/polisgo2020/search-tarival/web"
	"github.com/urfave/cli/v2"
)

var cfg config.Config

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	cfg = config.Load()
	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Str("log level", cfg.LogLevel)
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
			Action:  indexFunc,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "path",
					Aliases:  []string{"p"},
					Required: true,
					Usage:    "path to directory",
				},
			},
		},
		{
			Name:    "search",
			Aliases: []string{"s"},
			Usage:   "serching in directody with reverse index",
			Action:  searchFunc,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "index",
					Aliases:  []string{"i"},
					Required: true,
					Usage:    "path to reverse index",
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err)
	}
}

func indexFunc(c *cli.Context) error {

	path := c.String("path")

	if len(path) == 0 {
		log.Fatal().Err(errors.New("Path to folder not found"))
	}
	index, err := index.IndexingFolder(path)
	if err != nil {
		log.Fatal().Err(err)
	}
	output, err := json.Marshal(index)
	if err != nil {
		log.Fatal().Err(err)
	}
	if err := ioutil.WriteFile("index.json", output, 0666); err != nil {
		log.Fatal().Err(err)
	}
	return nil
}

func searchFunc(c *cli.Context) error {

	indexName := c.String("index")

	Index, err := index.ReadIndexJSON(indexName)
	if err != nil {
		log.Fatal().Err(err)
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
		log.Fatal().Err(err)
	}
	return nil
}
