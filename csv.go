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
	datasrc["CARDCATEGORY"] = []string{"NEW", "LEARN", "REVIEW"} // , "RELEARN"
	datasrc["CARDSTEP"] = []string{"ZERO", "ONE", "TWO", "THREE", "FOUR", "FIVE", "ENDING", "ENDED"}

	header := []string{"RATE", "ISDAILYTASK", "CARDCATEGORY", "CARDSTEP",
		"NEWCARDCATEGORY", "NEWCARDDUEDATE", "NEWCARDSTEP",
		"LEECHCOUNT", "UPDATELASTSEEN"}

	w := csv.NewWriter(os.Stdout)
	if err := w.Write(header); err != nil {
		log.Fatalln("error writing header to csv:", err)
	}

	for _, v1 := range datasrc["RATE"] {
		for _, v2 := range datasrc["ISDAILYTASK"] {
			for _, v4 := range datasrc["CARDCATEGORY"] {
				for _, v5 := range datasrc["CARDSTEP"] {
					line := []string{v1, v2, v4, v5}
					line = append(line, action(v1, v2, v4, v5)...)
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

// reaction indice
const (
	NEWCARDCATEGORY = iota
	NEWCARDDUEDATE
	NEWCARDSTEP
	LEECHCOUNT
	UPDATELASTSEEN
)

// SHOULDNOTHAPPEN is error
var SHOULDNOTHAPPEN = []string{"ERROR", "ERROR", "ERROR", "ERROR", "ERROR"}

func action(RATE, ISDAILYTASK, CARDCATEGORY, CURRENTCARDSTEP string) (act []string) {
	act = make([]string, 5)
	if RATE == "NOIDEA" && CARDCATEGORY != "NEW" {
		// if not a new card, and forgot, means it possiblly is a leech card
		act[LEECHCOUNT] = "PLUSONE"
	}
	if CARDCATEGORY == "NEW" {
		// anything NEW become LEARN after answered
		act[NEWCARDCATEGORY] = "LEARN"
	}

	switch RATE {
	case "WELLKNOWN":
		// not change CARDSTEP if its not due today
		if ISDAILYTASK == "YES" {
			act[NEWCARDSTEP], act[NEWCARDDUEDATE] = stepupDuedateAction(CURRENTCARDSTEP)
		} else {
			act[NEWCARDDUEDATE] = "EXTENDODUE"
		}
		switch act[NEWCARDSTEP] {
		case "ZERO", "ONE", "TWO":
		default:
			act[NEWCARDCATEGORY] = "REVIEW"
		}

	case "VAGUE", "NOIDEA":
		// review in 1 min
		act[NEWCARDDUEDATE] = "ADD1MINUTE"
		if CARDCATEGORY == "REVIEW" {
			act[NEWCARDCATEGORY] = "LEARN"
		}
		act[NEWCARDSTEP] = "ZERO"
	}

	return
}

func stepupDuedateAction(CURRENTCARDSTEP string) (string, string) {
	switch CURRENTCARDSTEP {
	case "ZERO":
		return "ONE", "ADD1MINUTE"
	case "ONE":
		return "TWO", "ADD15MINUTES"
	case "TWO":
		return "THREE", "ADD1DAY"
	case "THREE":
		return "FOUR", "ADD7DAYS"
	case "FOUR":
		return "FIVE", "ADD14DAYS"
	case "FIVE":
		return "ENDING", "ADD30DAYS"
	case "ENDING":
		return "ENDED", "ADD75DAYS"
	case "ENDED":
		return "ENDED", "ADD200DAYS"
	default:
		return "ERROR", "ERROR"
	}
}
