package main

import (
	"bufio"
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var dict = flag.String("dict", "./dict", "path of wordnet database files")
var tmpl = flag.String("tmpl", "./misc/wn-struct.sqlite", "path of sqlite database structure file")
var output = flag.String("output", "./wordnet.sqlite", "path to output sqlite db")

var (
	database                     *sql.DB
	sensestmt, exmstmt, wordstmt *sql.Stmt
)

func main() {
	flag.Parse()

	// tmpl db to output position
	copyFile(*tmpl, *output)

	// open sqlite database
	var err error
	database, err = sql.Open("sqlite3", "file:"+*output+"?cache=shared")
	if err != nil {
		log.Fatalln("failed to connect sqlite output file", err)
	}

	// prepare sql statements
	sensestmt, err = database.Prepare("INSERT INTO sense (gloss) VALUES (?)")
	if err != nil {
		log.Fatalln("failed to prepare insert into sense", err)
	}
	// prepare sql statements
	exmstmt, err = database.Prepare("INSERT INTO example (senseid, example) VALUES (?, ?)")
	if err != nil {
		log.Fatalln("failed to prepare insert into sense", err)
	}
	wordstmt, err = database.Prepare("INSERT INTO word (word, lex_id, ss_type, senseid, phrase) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalln("failed to prepare insert into word", err)
	}

	var categories = []string{"noun", "adj", "adv", "verb"}
	for _, v := range categories {
		process(v)
	}
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
			log.Println("[multi-glosses]", len(gloss), "file:", pos, "line:", lineid)
		}

		res, err := sensestmt.Exec(strings.Join(gloss, "; "))
		if err != nil {
			log.Println("file:", pos, "line:", lineid)
			log.Fatalln("fail to insert into sense gloss", err)
		}
		senseid, err := res.LastInsertId()
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
			isphrase := strings.Contains(word, " ")
			lexid := arr[5+i*2]
			res, err = wordstmt.Exec(word, lexid, ssType, senseid, isphrase)
			if err != nil {
				log.Fatalln(err, word)
			}
		}
		for _, v := range examples {
			res, err = exmstmt.Exec(senseid, v)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}

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
