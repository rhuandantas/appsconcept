package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type fizzbuzzParams struct {
	Int1  int    `json:"int1"`
	Int2  int    `json:"int2"`
	Limit int    `json:"limit"`
	Str1  string `json:"str1"`
	Str2  string `json:"str2"`
}

type statsEntry struct {
	Params fizzbuzzParams `json:"params"`
	Hits   int            `json:"hits"`
}

var (
	stats      = make(map[fizzbuzzParams]int)
	statsMutex sync.Mutex
)

func parseParams(r *http.Request) (fizzbuzzParams, error) {
	q := r.URL.Query()
	int1, err1 := strconv.Atoi(q.Get("int1"))
	int2, err2 := strconv.Atoi(q.Get("int2"))
	limit, err3 := strconv.Atoi(q.Get("limit"))
	str1 := q.Get("str1")
	str2 := q.Get("str2")

	if err1 != nil || err2 != nil || err3 != nil {
		return fizzbuzzParams{}, errors.New("parâmetros int1, int2 e limit devem ser inteiros")
	}
	if int1 <= 0 || int2 <= 0 || limit <= 0 || str1 == "" || str2 == "" {
		return fizzbuzzParams{}, errors.New("todos os parâmetros são obrigatórios e devem ser válidos")
	}
	return fizzbuzzParams{int1, int2, limit, str1, str2}, nil
}

func fizzbuzzLogic(params fizzbuzzParams) []string {
	result := make([]string, 0, params.Limit)
	for i := 1; i <= params.Limit; i++ {
		switch {
		case i%params.Int1 == 0 && i%params.Int2 == 0:
			result = append(result, params.Str1+params.Str2)
		case i%params.Int1 == 0:
			result = append(result, params.Str1)
		case i%params.Int2 == 0:
			result = append(result, params.Str2)
		default:
			result = append(result, strconv.Itoa(i))
		}
	}
	return result
}

func fizzbuzzHandler(w http.ResponseWriter, r *http.Request) {
	params, err := parseParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Erro de parâmetro: %v", err)
		return
	}

	statsMutex.Lock()
	stats[params]++
	statsMutex.Unlock()

	result := fizzbuzzLogic(params)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	statsMutex.Lock()
	defer statsMutex.Unlock()

	var maxEntry statsEntry
	for params, hits := range stats {
		if hits > maxEntry.Hits {
			maxEntry = statsEntry{params, hits}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(maxEntry)
	if err != nil {
		http.Error(w, "Erro ao codificar os dados de estatísticas", http.StatusInternalServerError)
		log.Printf("Erro ao codificar os dados de estatísticas: %v", err)
		return
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/fizzbuzz", fizzbuzzHandler)
	mux.HandleFunc("/stats", statsHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("Servidor rodando na porta 8080")
	log.Fatal(server.ListenAndServe())
}
