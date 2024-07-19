package types

type Database struct {
	Chirps struct {
		Num1 struct {
			ID   int    `json:"id"`
			Body string `json:"body"`
		} `json:"1"`
		Num2 struct {
			ID   int    `json:"id"`
			Body string `json:"body"`
		} `json:"2"`
	} `json:"chirps"`
}
