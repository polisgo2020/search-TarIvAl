package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/polisgo2020/search-tarival/index"
	"github.com/polisgo2020/search-tarival/search"

	"github.com/urfave/cli/v2"
)

const indexName = "index.json"

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
					Name:  "path",
					Usage: "path to directory",
				},
			},
		},
		{
			Name:    "search",
			Aliases: []string{"s"},
			Usage:   "serching in directody with reverse index",
			Action:  searchFunc,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func indexFunc(c *cli.Context) error {

	path := c.Args().Get(0)

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
	if err := ioutil.WriteFile(indexName, output, 0666); err != nil {
		log.Fatal(err)
	}
	return nil
}

func searchFunc(c *cli.Context) error {

	searchPhrase := c.Args().Get(0)

	if len(searchPhrase) == 0 {
		log.Fatal(errors.New("Search phrase not found"))
	}

	index, err := index.ReadIndexJSON(indexName)
	if err != nil {
		log.Fatal(err)
	}

	searchResult, err := search.Searching(index, searchPhrase)
	if err != nil {
		log.Fatal(err)
	}

	for i, result := range searchResult {
		fmt.Printf("%v) %v\n", i+1, result)
	}
	return nil
}
