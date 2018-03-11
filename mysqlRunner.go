package main

// runFiles runs each one of the files in the file list.
import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/morfien101/mysql-schema-updater/config"
)

// There are limitation that go introduces here.
// 1. We can only run one statement at a time.
// 2. In order to get the statements we split the sql files
//    on a ; followed directly but a \n (new line).
// This requies that the sql file are well formatted.
func runFiles(filelist []File, conf config.SQLConfig) {
	// - Connect to SQL
	// - Validate Connection
	// - Check DB exists create if asked to;
	//   set created flag to not bother version Check
	sqlConnection, newDB := connectToSQL(conf)
	// - Get version of DB if not new
	dbVersion := 0
	if newDB {
		// Create a new version table.
		_, err := sqlConnection.Exec(fmt.Sprintf("CREATE TABLE %s(id int not NULL PRIMARY KEY, file TEXT(500), checksum TEXT(32));", conf.SQLVersionTable()))
		if err != nil {
			log.Fatalf("There was an error while trying to create the %s table.", conf.SQLVersionTable())
		}
	}

	// Are there any version numbers available?
	var versionList int
	err := sqlConnection.QueryRow(fmt.Sprintf("SELECT id FROM %s LIMIT 1;", conf.SQLVersionTable())).Scan(&versionList)
	if err == sql.ErrNoRows {
		log.Printf("No version listings found in %s.%s. Assuming it is a new Database.", conf.SQLDB(), conf.SQLVersionTable())
	} else if err != nil {
		log.Fatalf("Failed to determin if there are version numbers available. Error: %s", err)
	} else {
		// Collect the Highed ID number from the table that holds them.
		err := sqlConnection.QueryRow(fmt.Sprintf("SELECT MAX(id) FROM %s;", conf.SQLVersionTable())).Scan(&dbVersion)
		if err != nil {
			log.Fatalf("Failed to collect the version numbers. Error: %s", err)
		}
	}
	// - check if we need to run any files
	if filelist[len(filelist)-1].Index <= dbVersion {
		log.Println("No update required.")
		log.Println("DB Version:", dbVersion)
		log.Println("Highest file version:", filelist[len(filelist)-1].Index)
	} else {
		// - run the files that need to be run.
		for _, filedata := range filelist {
			if filedata.Index <= dbVersion {
				continue
			}
			// We slurp the entire file into memory here.
			// Since this program would be used in a controlled manner and
			// well formatted text files are what we intend to use this with
			// should not cause an issue.
			query, _ := ioutil.ReadFile(filedata.Path)
			// split on a ; followed directly by a new line
			sq := strings.Split(string(query), ";\n")
			for _, stm := range sq {
				// We get an empty slice at the end of the sq slice
				if len(stm) == 0 {
					continue
				}
				// pretty log message telling the user what is about to be
				// executed for trouble shooting purposes.
				log.Println("Executing:", stm)
				// We don't really care for the output but rather errors.
				_, err := sqlConnection.Exec(stm)
				// Exit is SQL returns an error. Without context it would
				// be impossible to fix so this is the safest option.
				if err != nil {
					log.Fatal(err)
				}
			}
			// write version at the end of each file
			// We write the version and name of the file for auditing.
			_, err := sqlConnection.Exec(
				fmt.Sprintf(
					"INSERT INTO %s(id, file, checksum) VALUES (?, ?, ?);",
					conf.SQLVersionTable(),
				),
				filedata.Index,
				filedata.Path,
				filedata.md5Hash(),
			)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// connectToSQL connects to SQL and prepares our connection. We will
// if requested create the database in here also.
func connectToSQL(conf config.SQLConfig) (*sql.DB, bool) {
	// - Connect to SQL
	// Create the connection string to the root of the server.
	// Required if we need to create the Database first.
	connectionString := fmt.Sprintf("%s:%s@(%s:%d)/",
		conf.SQLUsername(),
		conf.SQLPassword(),
		conf.SQLHost(),
		conf.SQLPort(),
	)
	// Create the connecton. Go doesn't connect at this point, rahter
	// it waits till it needs to do something before making the connection.
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to the server. Error: %s", err)
	}
	// - Validate Connection
	// Make go create the connection to see if we are able to connect
	// to the server. Exit bad if we can't.
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to validate the connection. Error: %s", err)
	}
	// - Check DB exists create if asked to;
	//   set created flag to not bother version Check
	// Flag to return to say that the Database is new and all
	// SQL files need to be run against it.
	newDB := false
	// Query the DB server to see if the DB exists.
	q := fmt.Sprintf("SHOW DATABASES LIKE '%s';", conf.SQLDB())
	var holder string
	err = db.QueryRow(q).Scan(&holder)
	switch {
	case err == sql.ErrNoRows:
		// create the DB if asked to.
		if conf.CreateDB() {
			log.Println("DB not found creating.")
			_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s;", conf.SQLDB()))
			if err != nil {
				log.Fatal(err)
			} else {
				log.Print("DB Created")
			}
			newDB = true
		} else {
			log.Fatal("DB not found.")
		}
	case err != nil:
		log.Fatal(err)
	}

	// Tell the server to use the DB that we got from the user.
	_, err = db.Exec(fmt.Sprintf("USE %s;", conf.SQLDB()))
	if err != nil {
		log.Fatal(err)
	}

	// return back the db connection and the flag to tell if we created
	// the database.
	return db, newDB
}
