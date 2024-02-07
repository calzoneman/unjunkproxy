package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var reDiscordLink = regexp.MustCompile(`(cdn\.discordapp\.com|media\.discordapp\.net|images-ext-[12].discordapp.net)/.+`)

func HandleDiscordLink(w http.ResponseWriter, r *http.Request) {
	targetPath := reDiscordLink.FindString(r.URL.Path)
	if targetPath == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("unrecognized discord URL"))
		return
	}

	targetURL := fmt.Sprintf("https://%s", targetPath)
	log.Printf("proxying discord link %s?%s", targetURL, r.URL.RawQuery)
	preq, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		log.Printf("error creating request: %w", err)
		do500(w)
		return
	}

	preq.URL.RawQuery = r.URL.RawQuery

	client := &http.Client{}
	pres, err := client.Do(preq)
	if err != nil {
		log.Printf("error proxying request: %w", err)
		do500(w)
		return
	}
	defer pres.Body.Close()

	for k, vs := range pres.Header {
		if strings.ToLower(k) == "content-disposition" {
			continue
		}
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(pres.StatusCode)

	_, err = io.Copy(w, pres.Body)
	if err != nil {
		log.Printf("error copying response body: %w", err)
	}
}
