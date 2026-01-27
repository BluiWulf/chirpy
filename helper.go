package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func respondWithError(resW http.ResponseWriter, code int, msg string) {
	type errorResp struct {
		Error string `json:"error"`
	}
	respondWithJson(resW, code, errorResp{Error: msg})
}

func respondWithJson(resW http.ResponseWriter, code int, payload interface{}) {
	resW.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marhsalling JSON: %s", err)
		resW.WriteHeader(http.StatusInternalServerError)
		return
	}
	resW.WriteHeader(code)
	resW.Write(data)
}

func profanityCheck(body string) string {
	words := strings.Split(body, " ")
	for index, word := range words {
		lowerWord := strings.ToLower(word)
		if _, ok := profanity[lowerWord]; ok {
			words[index] = "****"
		}
	}
	return strings.Join(words, " ")
}
