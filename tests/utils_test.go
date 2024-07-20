package tests

import (
	"encoding/json"
	"github.com/mdwiltfong/chirpy/utils"
	"github.com/mdwiltfong/chirpy/utils/types"
	"os"
	"testing"
)

func cleanUp(t *testing.T) {
	delErr := os.Remove("../database/database.json")
	if delErr != nil {
		t.Fatal(delErr)
	}
}

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
	cleanUp(t)
}

func TestLoadDB(t *testing.T) {
	dbClient, _ := utils.NewDB("../database/database.json")
	dbStruct, err := dbClient.LoadDB()
	if err != nil {
		t.Fatal(err.Error())
	}
	if dbStruct.Chirps[1].Body != "This is the first chirp ever!" {
		t.Fatalf("Was expecting ")
	}
	cleanUp(t)
}

func TestWriteDB(t *testing.T) {
	dbClient, _ := utils.NewDB("../database/database.json")
	dbData, _ := dbClient.LoadDB()
	dbData.Chirps[3] = utils.Chirp{ID: 3, Body: "Test Chirp"}
	err := dbClient.WriteDB(dbData)
	if err != nil {
		t.Fatal(err.Error())
	}
	// Check that db file has new chirp
	dataBytes, err := os.ReadFile("../database/database.json")
	tempStruct := utils.DBStructure{}
	json.Unmarshal(dataBytes, &tempStruct)
	if tempStruct.Chirps[3].ID != 3 {
		t.Fatal("Incorrect ID stored")
	}
	if tempStruct.Chirps[3].Body != "Test Chirp" {
		t.Fatal("Incorrect Body stored")
	}
	cleanUp(t)
}
func TestEnsureDB(t *testing.T) {
	dbClient, _ := utils.NewDB("../database/database.json")
	os.Remove("../database/database.json")
	dbClient.EnsureDB()
	data, _ := os.ReadFile("../database/database.json")
	tempData := types.Database{}
	json.Unmarshal(data, &tempData)
	if tempData.Chirps.Num1.Body != "This is the first chirp ever!" {
		t.Fatal("There is no database.json file")
	}
	cleanUp(t)

}
