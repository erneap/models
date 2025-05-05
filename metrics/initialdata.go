package metrics

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/erneap/go-models/systemdata"
)

func InitialData() systemdata.SystemInfo {
	jsonFile, err := os.Open("/data/initial/initial.json")
	if err != nil {
		log.Println(err)
		jsonFile, err = os.Open("./initial.json")
		if err != nil {
			log.Println(err)
		}
	}

	log.Println("Opened Initial Data JSON File")
	defer jsonFile.Close()

	// read all the data of the jsonFile into a byteArray
	byteArray, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Println(err)
	}
	jsonString := string(byteArray)

	// set a variable for the system info
	var systemInfo systemdata.SystemInfo

	// unmarshall the json data into the system info struct
	err = json.Unmarshal([]byte(jsonString), &systemInfo)
	if err != nil {
		log.Println(err)
	}
	return systemInfo
}
