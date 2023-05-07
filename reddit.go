package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

var reRedditImage = regexp.MustCompile(`i\.redd\.it/.+`)

func HandleRedditImage(w http.ResponseWriter, r *http.Request) {
	targetPath := reRedditImage.FindString(r.URL.Path)
	if targetPath == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("unrecognized i.redd.it URL"))
		return
	}

	targetURL := fmt.Sprintf("https://%s", targetPath)
	log.Printf("proxying reddit image %s", targetURL)
	preq, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		log.Printf("error creating request: %w", err)
		do500(w)
		return
	}
	// fuck off with your JavaScript, an image URL ought to serve a damn image
	preq.Header.Set("Accept", "image/*")

	client := &http.Client{}
	pres, err := client.Do(preq)
	if err != nil {
		log.Printf("error proxying request: %w", err)
		do500(w)
		return
	}
	defer pres.Body.Close()

	w.WriteHeader(pres.StatusCode)
	for k, vs := range pres.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}

	_, err = io.Copy(w, pres.Body)
	if err != nil {
		log.Printf("error copying response body: %w", err)
	}
}
