package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/polisgo2020/search-tarival/index"
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
	interfaceListen := c.String("listen")

	index, err := index.ReadIndexJSON(indexName)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handleSearch)
	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		handleResult(w, r, index)
	})

	fmt.Println("Server started to listen at intterface ", interfaceListen)

	err = http.ListenAndServe(interfaceListen, nil)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func handleSearch(w http.ResponseWriter, r *http.Request) {

	query := r.FormValue("query")

	if len(query) == 0 {
		markup, _ := ioutil.ReadFile("web/templates/index.html")
		fmt.Fprintln(w, string(markup))
	} else {
		http.Redirect(w, r, "/result?query="+query, http.StatusFound)
	}
}

func handleResult(w http.ResponseWriter, r *http.Request, Index index.ReverseIndex) {

	query := r.FormValue("query")

	fmt.Printf("Get search phrase: %v\n", query)

	if len(query) == 0 {
		fmt.Fprintln(w, "Search phrase not found (r)")
		return
	}

	searchResult, err := Index.Searching(query)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}

	results := ""

	for i, result := range searchResult {
		results += fmt.Sprintf("<p>%v) %v</p>\n", i+1, result)
	}
	tmp, _ := template.ParseFiles("web/templates/index.html")

	req := struct {
		Results string
		Query   string
	}{
		Results: results,
		Query:   query,
	}

	tmp.Execute(w, req)
}
