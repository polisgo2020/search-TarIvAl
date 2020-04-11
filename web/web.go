package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/polisgo2020/search-tarival/index"
)

// HandleObject is parameter for ServerStart
type HandleObject struct {
	Address   string
	Func      func(w http.ResponseWriter, r *http.Request)
	FuncIndex func(w http.ResponseWriter, r *http.Request, index index.ReverseIndex)
	Index     index.ReverseIndex
}

// ServerStart is start the server at sliceHandleObj addresses, handle functions and index params
func ServerStart(listen string, timeout time.Duration, sliceHandleObj []HandleObject) error {
	mux := http.NewServeMux()

	server := http.Server{
		Addr:         listen,
		Handler:      mux,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	for _, handle := range sliceHandleObj {
		if handle.Func != nil {
			mux.HandleFunc(handle.Address, handle.Func)
			continue
		}
		if handle.FuncIndex != nil {
			mux.HandleFunc(handle.Address, func(w http.ResponseWriter, r *http.Request) {
				handle.FuncIndex(w, r, handle.Index)
			})
		}

	}

	fmt.Println("Server started to listen at interface ", listen)

	return server.ListenAndServe()
}
