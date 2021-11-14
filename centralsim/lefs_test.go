package centralsim

import (
	"log"
	"testing"
)

func TestLefsLoad(t *testing.T) {
	log.Printf("************************** Basic lefs loading test ....")
	lefs, err := Load("testdata/twoTwo.json")
	if err != nil {
		println("Couln't load th pn file !")
	}

	log := CreateLogger("0")
	lefs.ImprimeLefs(log)
}
