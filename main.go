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

	sliceHandleObj := []web.HandleObject{
		web.HandleObject{
			Address: "/",
			Func:    handleSearch,
		},
		web.HandleObject{
			Address:   "/result",
			FuncIndex: handleResult,
			Index:     Index,
		},
	}

	web.ServerStart(listen, 10*time.Second, sliceHandleObj)

	return nil
}

func handleResult(w http.ResponseWriter, r *http.Request, Index index.ReverseIndex) {
	query := r.FormValue("query")

	fmt.Printf("Get search phrase: %v\n", query)

	var results string

	searchResult, err := Index.Searching(query)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}

	if len(searchResult) == 0 {
		results = "Not found any result with your request"
	} else {
		for i, result := range searchResult {
			results += fmt.Sprintf("<p>%v) %v</p>\n", i+1, result)
		}
	}

	tmp, err := template.ParseFiles("web/templates/result.html")
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}

	tmpData := struct {
		Results string
		Query   string
	}{
		Results: results,
		Query:   query,
	}

	err = tmp.Execute(w, tmpData)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")

	if len(query) == 0 {
		markup, _ := template.ParseFiles("web/templates/index.html")
		markup.Execute(w, struct{}{})
	} else {
		http.Redirect(w, r, "/result?query="+query, http.StatusFound)
	}
}
