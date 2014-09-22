// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/st3fan/moz-tokenserver/tokenserver"
	"log"
	"net/http"
)

const (
	DEFAULT_API_PREFIX         = "/token"
	DEFAULT_API_LISTEN_ADDRESS = "0.0.0.0"
	DEFAULT_API_LISTEN_PORT    = 8123
)

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("1.0")) // TODO: How can we easily embed the git rev and tag in here?
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/version", VersionHandler)

	config := tokenserver.DefaultTokenServerConfig() // TODO: Get this from command line options

	_, err := tokenserver.SetupTokenServerRouter(router.PathPrefix(DEFAULT_API_PREFIX).Subrouter(), config)
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf("%s:%d", DEFAULT_API_LISTEN_ADDRESS, DEFAULT_API_LISTEN_PORT)
	log.Printf("Starting tokenserver server on http://%s", addr)
	http.Handle("/", router)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
