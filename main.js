
const sqlite3 = require('sqlite3').verbose();
const nlp = require('compromise');

const knownConj = ['FutureTense', 'PresentTense', 'PastTense', 'Gerund', 'Infinitive', 'Participle']

// open database in memory
let db = new sqlite3.Database('./wordnet.sqlite', (err) => {
  if (err) {
    return console.error(err.message);
  }
  console.log('Connected to the SQlite database.');
});
/*
db.run('CREATE TABLE "alias" ( "origin" TEXT, "alias" TEXT, "reason" TEXT)');
db.run('CREATE INDEX "idx_origin" ON "alias" ( "origin" )');
db.run('CREATE INDEX "idx_alias" ON "alias" ( "alias" )');
*/
/* test plural convertion
{
    let doc = nlp('child');
    doc.nouns().toPlural();
    console.log(doc.text());
}
*/


{
    let conj = nlp("fall").tag('Verb').verbs().conjugate()[0];
    Object.keys(conj).forEach((key) => {
        var val = conj[key];
        console.log(key + '->' + val);
        if (!knownConj.includes(key)) {
            console.log(key);
            process.exit(1);
        }
      });
    console.log(conj);
}


{ // tenses
    let sql = `SELECT DISTINCT word, phrase FROM word WHERE ss_type = "v"`;

    db.all(sql, [], (err, rows) => {
        if (err) {
            throw err;
        }
        rows.forEach((row) => {
            let word = row.word.replace('_', ' ').toLowerCase();
            var conjs = nlp(word)
            if (row.phrase < 1) {
                conjs = conjs.tag('Verb')
            }
            conjs = conjs.verbs().conjugate();
            conjs.forEach((conj) => {
                Object.keys(conj).forEach((key) => {
                    var val = conj[key];
                    if (val != word) {
                        console.log(word + '->' + val + ' @ ' + key);
                        db.run(`INSERT INTO alias(origin, alias, reason) VALUES(?,?,?)`, [row.word, val, key], function(err) {
                            if (err) {
                                console.log(err.message);
                                process.exit(1);
                                return
                            }
                            // get the last insert id
                            // console.log(`A row has been inserted with rowid ${this.lastID}`);
                        });
                    }
                  });
            });
        });
    });
}

{ // plural
    let sql = `SELECT DISTINCT word FROM word WHERE ss_type = "n"`;

    db.all(sql, [], (err, rows) => {
        if (err) {
            throw err;
        }
        rows.forEach((row) => {
            let word = row.word.replace('_', ' ').toLowerCase();
            var doc = nlp(word);
            if (row.phrase < 1) {
                doc = doc.tag('Noun');
            }
            doc = doc.nouns().toPlural();
            let plu = doc.text();
            if (plu != word) {
                console.log(row.word + ' -> ' + plu);
                db.run(`INSERT INTO alias(origin, alias, reason) VALUES(?,?,?)`, [row.word, plu, 'Plural'], function(err) {
                    if (err) {
                        console.log(err.message);
                        process.exit(1);
                        return
                    }
                    // get the last insert id
                    // console.log(`A row has been inserted with rowid ${this.lastID}`);
                });
            }
        });
    });
}
  
// close the database connection
db.close((err) => {
  if (err) {
    return console.error(err.message);
  }
  console.log('Close the database connection.');
});
