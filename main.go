package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/mdwiltfong/chirpy/utils"
	"github.com/mdwiltfong/chirpy/utils/types"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
	client, _ := utils.NewDB("database/database.json")
	apiCfg := apiConfig{
		filserverHits: 0,
		DBClient:      client,
		JWT_SECRET:    jwtSecret,
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
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.handleGetChirp)
	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	mux.HandleFunc("PUT /api/users", apiCfg.handleUpdateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefresh)
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
	JWT_SECRET    string
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
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 400, "Something went wrong")
	}
	respondWithJSON(w, 201, chirp)
}
func (cgf *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	// First, decode request to see if it's valid
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	decoder.Decode(&params)

	authHeader := r.Header.Get("Authorization")
	authHeaderArr := strings.SplitAfter(authHeader, " ")
	bearerToken := authHeaderArr[1]
	type MyCustomClaims struct {
		jwt.RegisteredClaims
	}
	parsedToken, err := jwt.ParseWithClaims(bearerToken, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cgf.JWT_SECRET), nil
	})

	if err != nil {
		respondWithError(w, 401, "Unauthorized request")
		return
	}
	userStrId, subjectErr := parsedToken.Claims.GetSubject()
	if subjectErr != nil {
		log.Print(subjectErr.Error())
		respondWithError(w, 503, "Server error")
		return
	}
	userId, _ := strconv.Atoi(userStrId)
	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), 10)
	updateUser := types.User{ID: userId, Email: params.Email, Password: hash}
	updatedUser, updatingErr := cgf.DBClient.UpdateUser(userId, updateUser)
	if updatingErr != nil {
		log.Print(updatingErr.Error())
		respondWithError(w, 401, "Unable to update user")
		return
	}
	updatedUser.Password = nil
	respondWithJSON(w, 200, updatedUser)

}
func (cgf *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	// First, decode request to see if it's valid
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	decoder.Decode(&params)
	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 503, "There was an issue creating the user")
		return
	}
	newUser, err := cgf.DBClient.CreateUsers(params.Email, hash)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 422, "There was an issue creating the user")
		return
	}
	respondWithJSON(w, 201, newUser)
}

func (cgf *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {

}

func (cgf *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type payload struct {
		ID           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	// First, decode request to see if it's valid
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	decodeErr := decoder.Decode(&params)
	if decodeErr != nil {
		log.Print(decodeErr.Error())
		respondWithError(w, 503, "Invalid payload")
	}
	// Search for user
	user, err := cgf.DBClient.GetUserByEmail(params.Email)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 401, err.Error())
		return
	}
	//Hash pw and store it
	pass := []byte(params.Password)
	hashErr := bcrypt.CompareHashAndPassword(user.Password, pass)
	if hashErr != nil {
		log.Print(hashErr.Error())
		respondWithError(w, 401, "Incorrect username or password")
		return
	}
	user.Password = nil
	tempExpiresAt := jwt.NewNumericDate(time.Now().UTC().Add(time.Hour))
	if params.ExpiresInSeconds != 0 {
		tempExpiresAt = jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(params.ExpiresInSeconds)))
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: tempExpiresAt,
		Subject:   strconv.Itoa(user.ID),
	})

	signedToken, signingErr := token.SignedString([]byte(cgf.JWT_SECRET))
	if signingErr != nil {
		log.Print(signingErr.Error())
		respondWithError(w, 503, "There was an issue logging in")
		return
	}
	user.Token = signedToken
	refreshToken, generateErr := cgf.DBClient.GenerateRefreshToken(user.ID)
	if generateErr != nil {
		respondWithError(w, 503, "There was an issue logging in")
	}
	result := payload{ID: user.ID, Email: user.Email, Token: user.Token, RefreshToken: refreshToken.Token}
	respondWithJSON(w, 200, result)
}
func (cgf *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	strChirpId := r.PathValue("chirpId")
	chirpId, err := strconv.Atoi(strChirpId)
	if err != nil {
		respondWithError(w, 500, "There was an issue with the provided chirp id")
	}
	dbStruct, err := cgf.DBClient.LoadDB()
	if err != nil {
		log.Print("Error: ", err.Error())
		respondWithError(w, 500, "Unable to read DB")
		return
	}

	dbChirp, err := findChirp(chirpId, dbStruct)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	respondWithJSON(w, 200, dbChirp)
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
func findChirp(chirpId int, dbstruct types.Database) (types.Chirp, error) {
	for i, chirp := range dbstruct.Chirps {
		if i == chirpId {
			return chirp, nil
		}
	}
	return types.Chirp{}, errors.New(fmt.Sprintf("Unable to find chirp: %v", chirpId))
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
	data, err := json.Marshal(payload)
	if err != nil {
		log.Print(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
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
