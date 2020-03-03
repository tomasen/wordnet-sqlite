package main

import (
	"bufio"
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/grokify/html-strip-tags-go"

	"github.com/flimzy/anki"
	_ "github.com/mattn/go-sqlite3"
)

var dict = flag.String("dict", "./dict", "path of wordnet database files")
var tmpl = flag.String("tmpl", "./misc/wn-struct.sqlite", "path of sqlite database structure file")
var output = flag.String("output", "./wordnet.sqlite", "path to output sqlite db")
var doanki = flag.Bool("anki", false, "proc anki tag")

var (
	database *sql.DB
	glossstmt, exmstmt, findwordstmt, updatesoundstmt,
	insertbookrefstmt, insertbookstmt,
	sense1stmt, word1stmt, updatepronuncstmt *sql.Stmt
)

func main() {
	flag.Parse()

	if *doanki {
		prepareDB()
		database.Exec("DELETE FROM book")
		database.Exec("DELETE FROM bookref")
		prepareStmts()
		procAnki()
		return
	}

	// tmpl db to output position
	copyFile(*tmpl, *output)

	prepareDB()

	prepareStmts()

	var categories = []string{"noun", "adj", "adv", "verb"}
	for _, v := range categories {
		process(v)
	}

	procAnki()
}
func prepareDB() {

	// open sqlite database
	var err error
	database, err = sql.Open("sqlite3", "file:"+*output+"?cache=shared")
	if err != nil {
		log.Fatalln("failed to connect sqlite output file", err)
	}
}

func prepareStmts() {
	var err error
	// prepare sql statements
	glossstmt, err = database.Prepare("INSERT INTO gloss (gloss) VALUES (?)")
	if err != nil {
		log.Fatalln("failed to prepare insert into gloss", err)
	}
	// prepare sql statements
	exmstmt, err = database.Prepare("INSERT INTO example (glossid, example) VALUES (?, ?)")
	if err != nil {
		log.Fatalln("failed to prepare insert into sense", err)
	}
	findwordstmt, err = database.Prepare("SELECT id FROM word WHERE word = ?")
	if err != nil {
		log.Fatalln("failed to prepare find word", err)
	}
	word1stmt, err = database.Prepare("INSERT INTO word (word, phrase) VALUES (?, ?)")
	if err != nil {
		log.Fatalln("failed to prepare insert into word", err)
	}
	sense1stmt, err = database.Prepare("INSERT INTO sense (word_id, lex_id, ss_type, glossid) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatalln("failed to prepare insert into sense", err)
	}
	updatepronuncstmt, err = database.Prepare("UPDATE word SET pronunc = ? WHERE word = ?")
	if err != nil {
		log.Fatalln("failed to prepare insert into word", err)
	}
	updatesoundstmt, err = database.Prepare("UPDATE word SET sound = ? WHERE word = ?")
	if err != nil {
		log.Fatalln("failed to prepare insert into word", err)
	}
	insertbookstmt, err = database.Prepare("INSERT INTO book (name, tag) VALUES (?, ?)")
	if err != nil {
		log.Fatalln("failed to prepare insert into book", err)
	}
	insertbookrefstmt, err = database.Prepare("INSERT INTO bookref (bookid, wordid) VALUES (?, ?)")
	if err != nil {
		log.Fatalln("failed to prepare insert into word", err)
	}

}
func procAnki() {

	procAnkiData("COCAEnglish10000", "COCA", "COCA 10000", 0, 1, -2)
	// procAnkiData("4000_Essential_English_Words_all_books_en-en", "BASIC", "ESSENTIAL 400", 0, 7, 2)
	// procAnkiData("TOEFL", "TOEFL", "TOEFL", 0, -1, -2)
	// procAnkiData("SAT3500", "SAT1", "SAT 3500", 0, -1, -1)
	// procAnkiData("SAT6000", "SAT2", "SAT 6000", 1, -1, -1)
	// procAnkiData("GRE", "GRE", "GRE", 0, 1, 7)
}

func procAnkiData(fname, tag, name string, wordidx, proncid, soundid int) {
	// word with pronunciation from anki
	p, err := anki.ReadFile("./misc/" + fname + ".apkg")
	if err != nil {
		log.Fatalln(err)
	}

	res, err := insertbookstmt.Exec(name, tag)
	if err != nil {
		log.Fatalln("fail to insert into book", err)
	}
	bookid, err := res.LastInsertId()
	if err != nil {
		log.Fatalln("fail to get book id", err)
	}
	var cword, cpronc, csound int
	notes, err := p.Notes()
	for notes.Next() {
		n, err := notes.Note()
		if err != nil {
			log.Fatalln(err)
		}
		// for k, v := range n.FieldValues {
		// 	log.Println(k, v)
		// }
		// return
		word := n.FieldValues[wordidx]
		soundfile := ""
		if soundid == -2 {
			var sf = regexp.MustCompile(`.*(\[sound\:.+\])`)
			p := sf.FindStringSubmatch(word)
			if len(p) < 2 {
				log.Fatalln("error get sound", word)
			}
			soundfile = p[1]
			soundfile = strings.ToLower(soundfile)
			word = strings.Replace(word, soundfile, "", -1)
			word = strings.TrimSpace(word)
		}
		word = strip.StripTags(word)
		wordid := findWordID(word)
		if wordid < 0 {
			log.Println("WARN: word not found", word, n.FieldValues[wordidx])
			continue
		}
		if proncid >= 0 {
			pronc := n.FieldValues[proncid]
			res, err := updatepronuncstmt.Exec(pronc, word)
			if err != nil {
				log.Fatalln("upable to update pronunciation to word")
			}
			m, err := res.RowsAffected()
			if err != nil || m < 1 {
				log.Fatalln("upable to update pronunciation to word 2")
			}
			cpronc++
		}
		if soundid >= 0 {
			soundfile = n.FieldValues[soundid]
		}
		if len(soundfile) > 0 {
			sound := getAnkiSoundFile(p, soundfile)
			if len(sound) > 0 {
				res, err := updatesoundstmt.Exec(sound, word)
				if err != nil {
					log.Fatalln("upable to update sound to word")
				}
				m, err := res.RowsAffected()
				if err != nil || m < 1 {
					log.Fatalln("upable to update sound to word 2")
				}
				csound++
			}
		}
		res, err = insertbookrefstmt.Exec(bookid, wordid)
		if err != nil {
			log.Println("upable to insert bookref", err, bookid, wordid, word, n.FieldValues[wordidx])
		} else {
			m, err := res.RowsAffected()
			if err != nil || m < 1 {
				log.Fatalln("upable to insert bookref 2", n.FieldValues[wordidx])
			}
			cword++
		}
	}
	log.Println("done", cword, "words,", csound, "sound,", cpronc, "prounc")
}

func process(pos string) {
	// open data file
	datafile, err := os.Open(path.Join(*dict, "data."+pos))
	if err != nil {
		log.Fatalln("failed to open data:", pos, "error", err)
	}
	defer datafile.Close()

	lineid := 0
	scanner := bufio.NewScanner(datafile)
	for scanner.Scan() {
		line := scanner.Text()
		lineid++
		if strings.HasPrefix(line, " ") {
			continue
		}
		// not a comment
		// synset_offset  lex_filenum  ss_type  w_cnt  word \
		// lex_id  [word  lex_id...]  p_cnt  [ptr...]  [frames...]  |   gloss
		// lex_id: One digit hexadecimal integer that, when appended onto lemma ,
		// uniquely identifies a sense within a lexicographer file. lex_id numbers
		// usually start with 0 , and are incremented as additional senses of the
		// word are added to the same file, although there is no requirement that the
		// numbers be consecutive or begin with 0 . Note that a value of 0 is the default,
		// and therefore is not present in lexicographer files.
		arr := strings.Split(line, "|")
		if len(arr) != 2 {
			log.Println("file:", pos, "line:", lineid)
			log.Fatalln("unrecogenized gloss")
		}
		gloss := []string{}
		examples := []string{}
		for _, v := range strings.Split(arr[1], ";") {
			v = strings.TrimSpace(v)
			if strings.HasPrefix(v, "\"") {
				examples = append(examples, strings.Trim(v, "\""))
			} else {
				gloss = append(gloss, v)
			}
		}
		if len(gloss) > 1 {
			// log.Println("[multi-glosses]", len(gloss), "file:", pos, "line:", lineid)
		}

		res, err := glossstmt.Exec(strings.Join(gloss, "; "))
		if err != nil {
			log.Println("file:", pos, "line:", lineid)
			log.Fatalln("fail to insert into sense gloss", err)
		}
		glossid, err := res.LastInsertId()
		if err != nil {
			log.Println("file:", pos, "line:", lineid)
			log.Fatalln("fail to retrieve last insert id", err)
		}

		arr = strings.Split(arr[0], " ")
		ssType := arr[2]
		wCnt, err := strconv.ParseInt(arr[3], 16, 32)
		if err != nil {
			log.Println("file:", pos, "line:", lineid)
			log.Fatalln("error:", err)
		}
		for i := 0; i < int(wCnt); i++ {
			word := strings.Replace(arr[4+i*2], "_", " ", -1)
			wordID := ensureWordID(word)
			lexid, err := strconv.ParseInt(arr[5+i*2], 16, 32)
			if err != nil {
				log.Fatalln(err, arr[5+i*2])
			}
			res, err = sense1stmt.Exec(wordID, lexid, ssType, glossid)
			if err != nil {
				log.Fatalln(err, word)
			}
		}
		for _, v := range examples {
			res, err = exmstmt.Exec(glossid, v)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
}

func getAnkiSoundFile(apkg *anki.Apkg, mediafield string) []byte {
	if len(mediafield) > 0 {
		mediafield = strings.Trim(mediafield, "[]")
		fe := strings.Split(mediafield, ":")
		if fe[0] == "sound" && strings.HasSuffix(fe[1], ".mp3") {
			mediafield = fe[1]
		} else {
			log.Fatalln(mediafield)
		}
		b, err := apkg.ReadMediaFile(mediafield)
		if err != nil {
			log.Println(err)
			return nil
		}
		return b
	}
	return nil
}

func findWordID(word string) (wordid int64) {
	err := findwordstmt.QueryRow(word).Scan(&wordid)
	if err != nil {
		// log.Println("WARN: word not found", word)
		return -1
	}
	return
}

func ensureWordID(word string) (wordid int64) {
	err := findwordstmt.QueryRow(word).Scan(&wordid)
	if err != nil {
		isphrase := strings.Contains(word, " ")
		res, err := word1stmt.Exec(word, isphrase)
		if err != nil {
			log.Fatalln("fail to insert word")
		}
		wordid, err = res.LastInsertId()
		if err != nil {
			log.Fatalln("fail to get last insert word id")
		}
	}
	return
}

func copyFile(sourceFile, destinationFile string) error {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		log.Fatalln("failed to open struct sqlite template", err)
		return err
	}

	err = ioutil.WriteFile(destinationFile, input, 0644)
	if err != nil {
		log.Fatalln("Error", err, "creating", destinationFile)
		return err
	}

	return nil
}
