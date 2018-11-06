package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

// single-threaded for now
var globalAnalyzer = struct {
	sync.Mutex
	*analyzer
}{analyzer: newAnalyzer()}

func init() {
	http.HandleFunc("/findreferences", handleFindReferences)
	http.HandleFunc("/gotodefinition", handleGoToDefinition)
	http.HandleFunc("/hover", handleHover)
}

type (
	goToDefinitionRequest struct {
		FilePath    string
		Row, Column int
	}
	goToDefinitionResponse struct {
		OK          bool
		FilePath    string
		Row, Column int
	}
)

func handleGoToDefinition(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "missing body", 400)
		return
	}

	var req goToDefinitionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("invalid request", err)
		http.Error(w, err.Error(), 400)
		return
	}

	globalAnalyzer.Lock()
	dst, err := globalAnalyzer.findDefinition(req.FilePath, nil, req.Row, req.Column)
	globalAnalyzer.Unlock()

	var res goToDefinitionResponse
	if err != nil {
		log.Println("failed to find definition", err)
		res.OK = false
	} else {
		res.OK = true
		res.FilePath = dst.Filename
		res.Row = dst.Line
		res.Column = dst.Column
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println("failed to encode response", err)
		return
	}
}

type (
	hoverRequest struct {
		FilePath    string
		Row, Column int
	}
	hoverResponse struct {
		OK   bool
		Text string
	}
)

func handleHover(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "missing body", 400)
		return
	}

	var req hoverRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("invalid request", err)
		http.Error(w, err.Error(), 400)
		return
	}

	globalAnalyzer.Lock()
	docs, err := globalAnalyzer.getDocs(req.FilePath, nil, req.Row, req.Column)
	globalAnalyzer.Unlock()

	var res hoverResponse
	if err != nil {
		log.Println("failed to get docs", err)
		res.OK = false
	} else {
		res.OK = true
		res.Text = docs
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println("failed to encode response", err)
		return
	}
}

type (
	findReferencesRequest struct {
		FilePath    string
		Row, Column int
	}
	findReferencesResponse struct {
		OK         bool
		References []reference
	}
	reference struct {
		FilePath    string
		Row, Column int
	}
)

func handleFindReferences(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "missing body", 400)
		return
	}

	var req findReferencesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("invalid request", err)
		http.Error(w, err.Error(), 400)
		return
	}

	globalAnalyzer.Lock()
	references, err := globalAnalyzer.findReferences(req.FilePath, nil, req.Row, req.Column)
	globalAnalyzer.Unlock()

	var res findReferencesResponse
	if err != nil {
		log.Println("failed to find references", err)
		res.OK = false
	} else {
		res.OK = true
		for _, ref := range references {
			res.References = append(res.References, reference{
				FilePath: ref.Filename,
				Row:      ref.Line,
				Column:   ref.Column,
			})
		}
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println("failed to encode response", err)
		return
	}
}
