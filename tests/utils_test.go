package tests

import (
	"encoding/json"
	"github.com/mdwiltfong/chirpy/utils"
	"github.com/mdwiltfong/chirpy/utils/types"
	"os"
	"testing"
)

func TestDbFunctions(t *testing.T) {
	dbClient, err := utils.NewDB("../database/database.json")
	if err != nil {
		t.Fatal("There was an issue creating a DB connection: ", err)
	}
	if dbClient.Path != "../database/database.json" {
		t.Fatal("Db path is incorrect")
	}
	// Check that a database.json file was actually made
	data, _ := os.ReadFile("../database/database.json")
	tempData := types.Database{}
	json.Unmarshal(data, &tempData)
	if tempData.Chirps.Num1.Body != "This is the first chirp ever!" {
		t.Fatal("There is no database.json file")
	}
	// Clean up

	delErr := os.Remove("../database/database.json")
	if delErr != nil {
		t.Fatal(delErr)
	}

}
