package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
)

var reImgurAlbum = regexp.MustCompile(`imgur.com/(?:a|gallery)/([a-zA-Z0-9]+)`)
var reImgurSingle = regexp.MustCompile(`imgur.com/([a-zA-Z90-9]+)`)

var imgurAlbumPage = template.Must(template.New("imgur_album").Parse(`
<!doctype html>
<html>
	<head>
		<title>Imgur Album {{.Title}}</title>
		<style type="text/css">
		* {
			box-sizing: border-box;
		}

		body {
			background-color: #f0f0f0;
			max-width: 640px;
			margin: auto;
		}

		img {
			width: 100%;
		}
		</style>
	</head>
	<body>
		{{range .Media}}
			<a href="{{.URL}}" rel="noopener noreferer"><img src="{{.URL}}"></a>
			<hr>
		{{end}}
	</body>
</html>
`))

type imgurAlbumResults struct {
	Title string       `json:"title"`
	Media []imgurMedia `json:"media"`
}

type imgurMedia struct {
	URL string `json:"url"`
}

// TODO: consolidate common logic for image proxying
func HandleImgurSingle(w http.ResponseWriter, r *http.Request) {
	match := reImgurSingle.FindStringSubmatch(r.URL.Path)
	if len(match) < 2 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("unrecognized imgur image"))
		return
	}

	targetURL := fmt.Sprintf("https://i.imgur.com/%s.png", match[1])
	log.Printf("proxying imgur image %s", targetURL)
	preq, err := http.NewRequest("GET", targetURL, nil)
	preq.Header.Set("User-Agent", r.Header.Get("User-Agent"))
	// fuck OFF
	preq.Header.Set("Accept", "image/*")
	if err != nil {
		log.Printf("error creating request: %w", err)
		do500(w)
		return
	}

	client := &http.Client{}
	pres, err := client.Do(preq)
	if err != nil {
		log.Printf("error proxying request: %w", err)
		do500(w)
		return
	}
	defer pres.Body.Close()

	for k, vs := range pres.Header {
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

func HandleImgurAlbum(w http.ResponseWriter, r *http.Request) {
	match := reImgurAlbum.FindStringSubmatch(r.URL.Path)
	if len(match) < 2 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("unrecognized imgur album"))
		return
	}

	albumURL := fmt.Sprintf("https://api.imgur.com/post/v1/albums/%s?client_id=546c25a59c58ad7&include=media", match[1])
	log.Printf("proxying imgur album %s", match[1])
	preq, err := http.NewRequest("GET", albumURL, nil)
	if err != nil {
		log.Printf("error creating request: %w", err)
		do500(w)
		return
	}

	client := &http.Client{}
	pres, err := client.Do(preq)
	if err != nil {
		log.Printf("error proxying request: %w", err)
		do500(w)
		return
	} else if pres.StatusCode != http.StatusOK {
		log.Printf("error proxying request: HTTP %d", pres.StatusCode)
		do500(w)
		return
	}
	defer pres.Body.Close()

	var album imgurAlbumResults
	if err = json.NewDecoder(pres.Body).Decode(&album); err != nil {
		log.Printf("error proxying request: %w", err)
		do500(w)
	}

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err = imgurAlbumPage.Execute(w, &album)
	if err != nil {
		log.Printf("error executing template: %w", err)
	}
}
