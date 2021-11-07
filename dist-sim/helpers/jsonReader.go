// source: https://tutorialedge.net/golang/parsing-json-with-golang/
package helpers

import (
	"encoding/json"
	"io/ioutil"
	"petrisim/models"
)

// const path = "/home/a846866/Documents/"
const path = "/home/freedy/Documents/master/RedesYDistribuidos/MiniproyectoSD/petri-net-sim/dist-sim/"

func readFile(fileName string) []byte {
	data, err := ioutil.ReadFile(path + fileName)
	if err != nil {
		panic(err)
	}

	return data
}

func ReadNetConfig(fileName string) []models.ProcessInfo {
	data := readFile(fileName)

	var myJson []models.ProcessInfo
	err := json.Unmarshal(data, &myJson)
	if err != nil {
		panic(err)
	}

	return myJson
}
