package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/polisgo2020/search-tarival/index"
	"github.com/polisgo2020/search-tarival/web"
	"github.com/urfave/cli/v2"
)

func main() {
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
				&cli.StringFlag{
					Name:     "listen",
					Aliases:  []string{"l"},
					Required: true,
					Usage:    "interface for listening",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func indexFunc(c *cli.Context) error {

	path := c.String("path")

	if len(path) == 0 {
		log.Fatal(errors.New("Path to folder not found"))
	}
	index, err := index.IndexingFolder(path)
	if err != nil {
		log.Fatal(err)
	}
	output, err := json.Marshal(index)
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("index.json", output, 0666); err != nil {
		log.Fatal(err)
	}
	return nil
}

func searchFunc(c *cli.Context) error {
	indexName := c.String("index")
	listen := c.String("listen")

	Index, err := index.ReadIndexJSON(indexName)
	if err != nil {
		log.Fatal(err)
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

	if err = web.ServerStart(listen, 10*time.Second, handleObjs); err != nil {
		log.Fatal(err)
	}
	return nil
}
