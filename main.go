package main

import (
	"encoding/json"
	"fmt"
	"github.com/mdwiltfong/chirpy/utils"
	"log"
	"net/http"
	"strings"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
	client, _ := utils.NewDB("database/database.json")
	apiCfg := apiConfig{
		filserverHits: 0,
		DBClient:      client,
	}
	mux.Handle("/app/*", http.StripPrefix("/app",
		apiCfg.middlewareMetricInc(http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerAdminMetrics)
	mux.HandleFunc("/api/reset", apiCfg.handleReset)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValideateChirp)
	mux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirps)
	mux.HandleFunc("GET /api/chirps", apiCfg.handleReadChirps)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

type apiConfig struct {
	filserverHits int
	DBClient      *utils.DataBaseClient
}

func (cgf *apiConfig) handleCreateChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"Body"`
	}
	// First, decode request to see if it's valid
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	decoder.Decode(&params)
	chirp, err := cgf.DBClient.CreateChirp(params.Body)
	if err != nil {
		//It's not valid, so we have to prepare a response
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 400, "Something went wrong")
		return
	}
	respondWithJSON(w, 201, chirp)
}
func (cgf *apiConfig) handleReadChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cgf.DBClient.GetChirps()
	if err != nil {
		respondWithError(w, 503, err.Error())
	}
	respondWithJSON(w, 200, chirps)
}
func (cgf *apiConfig) middlewareMetricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cgf.filserverHits++
		next.ServeHTTP(w, r)
	})
}
func (cgf *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	body := fmt.Sprintf("Hits: %v", cgf.filserverHits)
	w.Write([]byte(body))
}

func (cgf *apiConfig) handlerAdminMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	body := fmt.Sprintf(`<html>
<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>
</html>`, cgf.filserverHits)
	w.Write([]byte(body))
}
func (cgf *apiConfig) handlerValideateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"Body"`
	}
	// First, decode request to see if it's valid
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {

		//It's not valid, so we have to prepare a response
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 400, "Something went wrong")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	respondWithCleanedBody(w, params.Body)
}
func (cgf *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cgf.filserverHits = 0
	body := fmt.Sprintf("Hits: %v", cgf.filserverHits)
	w.Write([]byte(body))
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func respondWithCleanedBody(w http.ResponseWriter, chirp string) {
	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(chirp, " ")
	for i := 0; i < len(words); i++ {
		lowerCaseWord := strings.ToLower(words[i])
		if _, ok := profaneWords[lowerCaseWord]; ok {
			words[i] = "****"
		}
	}

	type cleanedResponse struct {
		CleanBody string `json:"cleaned_body"`
	}
	cleanResp := cleanedResponse{
		CleanBody: strings.Join(words, " "),
	}

	respondWithJSON(w, 200, cleanResp)

}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	data, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Println("Outgoing Data: ", data)
	w.Write(data)
	return
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	type returnVal struct {
		Error string `json:"error"`
	}
	respBody := returnVal{
		Error: msg,
	}
	data, _ := json.Marshal(respBody)
	w.Write(data)
	return
}
