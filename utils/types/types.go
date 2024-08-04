package types

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}
type User struct {
	ID             int    `json:"id"`
	Email          string `json:"email"`
	Password       []byte `json:"password,omitempty"`
	Token          string `json:"token,omitempty"`
	RefreshTokenId int    `json:"refresh_token_id,omitempty"`
}
type RefreshToken struct {
	ID      int    `json:"id"`
	UserId  int    `json:"userId"`
	Token   string `json:"token"`
	IsValid bool   `json:"is_valid"`
}
type Database struct {
	Chirps        map[int]Chirp        `json:"chirps"`
	Users         map[int]User         `json:"users"`
	RefreshTokens map[int]RefreshToken `json:"refresh_tokens"`
}

type CustomClaims struct {
	Email     string `json:"email"`
	Password  []byte `json:"password"`
	ExpiresAt string `json:"expires_at"`
}
