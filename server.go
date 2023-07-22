package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8000", "http listen address")

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	mux := http.NewServeMux()
	mux.HandleFunc("/i.redd.it/", HandleRedditImage)
	mux.HandleFunc("/cdn.discordapp.com/", HandleDiscordLink)
	mux.HandleFunc("/media.discordapp.net/", HandleDiscordLink)
	mux.HandleFunc("/images-ext-1.discordapp.net/", HandleDiscordLink)
	mux.HandleFunc("/images-ext-2.discordapp.net/", HandleDiscordLink)
	mux.HandleFunc("/imgur.com/a/", HandleImgurAlbum)
	mux.HandleFunc("/imgur.com/", HandleImgurSingle)

	srv := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	log.Printf("starting server on %s", srv.Addr)

	log.Fatal(srv.ListenAndServe())
}
