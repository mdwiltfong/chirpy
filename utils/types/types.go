package types

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}
type Database struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}
