package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

	// indexFile := c.Args().Get(0)
	port := c.Args().Get(1)

	http.HandleFunc("/", handleSearch)

	fmt.Println("Server started at port ", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func handleSearch(w http.ResponseWriter, r *http.Request) {

	searchPhrase := r.URL.Query().Get("searchPhrase")

	if len(searchPhrase) == 0 {
		fmt.Fprintln(w, "Search phrase not found")
		return
	}

	index, err := index.ReadIndexJSON(indexName)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}

	searchResult, err := search.Searching(index, searchPhrase)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}

	for i, result := range searchResult {
		fmt.Fprintf(w, "<p>%v) %v</p>\n", i+1, result)
	}
}
