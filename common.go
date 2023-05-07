package main

import (
	"net/http"
)

func do500(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("internal error"))
	return
}
