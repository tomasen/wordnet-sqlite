package main

import (
	"encoding/csv"
	"log"
	"os"
)

func main() {
	datasrc := make(map[string][]string, 1)
	datasrc["RATE"] = []string{"WELLKNOWN", "VAGUE", "NOIDEA"}
	datasrc["ISDAILYTASK"] = []string{"YES", "NO"}
	datasrc["CURRENTDAILYSTEP"] = []string{"ZERO", "ONE", "ENDED"}
	datasrc["CARDCATEGORY"] = []string{"NEW", "LEARN", "REVIEW"}

	header := []string{"RATE", "ISDAILYTASK",
		"CURRENTDAILYSTEP", "CARDCATEGORY",
		"NEWCARDCATEGORY", "NEWDUEDATEWORDCARD",
		"NEWDAILYSTEP", "NEWDUETIMETODAY", "LEECHCOUNT"}
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(header); err != nil {
		log.Fatalln("error writing header to csv:", err)
	}

	for _, v1 := range datasrc["RATE"] {
		for _, v2 := range datasrc["ISDAILYTASK"] {
			for _, v3 := range datasrc["CURRENTDAILYSTEP"] {
				for _, v4 := range datasrc["CARDCATEGORY"] {
					line := []string{v1, v2, v3, v4}
					line = append(line, action(v1, v2, v3, v4)...)
					if err := w.Write(line); err != nil {
						log.Fatalln("error writing line to csv:", err)
					}
				}
			}
		}
	}
	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	log.Println("done")
}

const (
	NEWCARDCATEGORY = iota
	NEWDUEDATEWORDCARD
	NEWDAILYSTEP
	NEWDUETIMETODAY
	LEECHCOUNT
)

func action(RATE, ISDAILYTASK, CURRENTDAILYSTEP, CARDCATEGORY string) (act []string) {
	act = make([]string, 5)
	if RATE == "NOIDEA" && CARDCATEGORY != "NEW" {
		act[LEECHCOUNT] = "PLUSONE"
	}

	if ISDAILYTASK == "YES" {

	}

	return
}
