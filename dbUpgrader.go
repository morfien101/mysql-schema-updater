// Steps to achive
// *1. create flags for the following:
// - mysql host
// - mysql port
// - mysql username
// - mysql password
// - mysql database
// - mysql new database flag
// - mysql version table
// - files location
// - help
// *2. gather the files and sort into a numbered order
// *3. connect to the MySQL Server
// *4. lookup the version (Table could not exist)
// *5. apply scripts in required order and write the version and file
//    to the version table

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Setup the flags that we need.
var (
	sqlHost         = flag.String("sqlhost", "localhost", "Host of the SQL server")
	sqlPort         = flag.Int("sqlport", 3306, "Port to connect to the SQL server on.")
	sqlUserName     = flag.String("sqlusername", "root", "Username to connect to the SQL server with")
	sqlPassword     = flag.String("sqlpassword", "root", "Password to connect to the SQL server with")
	sqldb           = flag.String("sqldb", "test", "Name of the DB that we should be connecting to.")
	sqlVersionTable = flag.String("sqlversion-table", "version", "Name of the DB table that keeps version information")
	scriptPath      = flag.String("scripts-path", "/tmp/scripts", "Directory containing the upgrade scripts to run.")
	createDB        = flag.Bool("create-db", false, "Create the Database if it does not exist.")
	help            = flag.Bool("help", false, "Display Help message.")
)

// File holds data about each file we have
type File struct {
	Index int
	Path  string
}

// String allows the type to be output as a sting native value.
func (f File) String() string {
	return fmt.Sprintf("%d - %s", f.Index, f.Path)
}

// Implement the sorter and functions
type byIndex []File

func (a byIndex) Len() int           { return len(a) }
func (a byIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byIndex) Less(i, j int) bool { return a[i].Index < a[j].Index }

// main is what is called when we start execution
func main() {
	// Deal with the flags
	flag.Parse()
	// Check that if there are flags and if the help flag was set
	if *help == true || len(os.Args) < 2 {
		displayHelp()
		os.Exit(0)
	}

	// Gather files.
	files := getFiles(scriptPath)
	// Check for duplication in file numbering
	validateOrder(files)
	// Execute the SQL in the files.
	runFiles(files)

}

// displayHelp Shows the help message
func displayHelp() {
	s := `This script is used to upgrade sql servers.
It takes upgrade scripts that are numbers and applies then in the
numbered order. The numbers must appear at the begining of the file.
The can have gaps in them and can also have leading zeros if that
helps make things look better.

The database must have a table called version that will accomodate
the version number the database is on and also the names of the
files that have been run on the verson.

CREATE TABLE version(id int not NULL PRIMARY KEY, file TEXT(500));
`
	fmt.Println(s)
	flag.PrintDefaults()
}

// getFiles gets the file names from disk, sorts them
// and returns a slice of type Files. This will have
// the file name and its digested id.
func getFiles(path *string) []File {
	filesSlice := gatherFiles(*path)
	return sortFiles(filesSlice)
}

// gatherFiles will get the all the .sql files from the
// passed in location.
func gatherFiles(files string) (sqlfiles []string) {
	if _, err := os.Stat(files); err == nil {
		sqlfiles, err = filepath.Glob(fmt.Sprintf("%s/*.sql", files))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatalf("%s doesn't appear to be a valid path.", files)
	}
	return
}

// sortFiles will return back a slie of type File that is
// sorted in accending order.
func sortFiles(fileslice []string) []File {
	sortedFiles := makeFileList(fileslice)
	sort.Sort(byIndex(sortedFiles))
	return sortedFiles
}

// makeFileList will return a slice of type File that is
// not sorted.
func makeFileList(fileslice []string) (list []File) {
	//[["1","001filename.sql"] ["2", 02.filename.sql]]
	list = make([]File, len(fileslice))
	for i, f := range fileslice {
		list[i] = *digestFileName(f)
	}
	return
}

// digestFileName will return a pointer to a File struct.
// The file struct will have the Index set which we get
// from regex capturing the digits are the start of the
// the file. This is then cast to a int.
func digestFileName(f string) *File {
	reStartDigit := regexp.MustCompile("^([[:digit:]]+).*\\.sql")
	reStartZero := regexp.MustCompile("^0")
	reZero := regexp.MustCompile("^0+([1-9][0-9]*)")

	data := new(File)
	data.Path = f
	sp := strings.Split(f, "/")
	fname := sp[len(sp)-1]
	sd := reStartDigit.FindStringSubmatch(fname)
	var number string
	if reStartZero.MatchString(sd[1]) == true {
		number = reZero.FindStringSubmatch(sd[1])[1]
	} else {
		number = sd[1]
	}
	var err error
	data.Index, err = strconv.Atoi(number)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

// validateOrder looks for duplications in numbers that
// that the file names present. This would cause a failure
// as we can not determine the correct processing order.
// Exit bad in this situation.
func validateOrder(filelist []File) {
	for i := range filelist {
		for _, l := range filelist {
			if filelist[i].Path == l.Path {
				continue
			} else {
				if filelist[i].Index == l.Index {
					log.Fatalf("There appears to be duplicate numbered files.\nCan't determine safe order.\n%s\n%s", filelist[i].Path, l.Path)
				}
			}
		}
	}
}

// runFiles runs each one of the files in the file list.
// There are limitation that go introduces here.
// 1. We can only run one statement at a time.
// 2. In order to get the statements we split the sql files
//    on a ; followed directly but a \n (new line).
// This requies that the sql file are well formatted.
func runFiles(filelist []File) {
	// - Connect to SQL
	// - Validate Connection
	// - Check DB exists create if asked to;
	//   set created flag to not bother version Check
	sqlConnection, newDB := connectToSQL()
	// - Get version of DB if not new
	dbVersion := 0
	if newDB == false {
		// Collect the Highed ID number from the table that holds them.
		err := sqlConnection.QueryRow(fmt.Sprintf("SELECT MAX(id) FROM %s;", *sqlVersionTable)).Scan(&dbVersion)
		if err != nil {
			log.Fatal(err)
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
			_, err := sqlConnection.Exec(fmt.Sprintf("INSERT INTO %s(id, file) VALUES (?, ?);", *sqlVersionTable), filedata.Index, filedata.Path)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// connectToSQL connects to SQL and prepares our connection. We will
// if requested create the database in here also.
func connectToSQL() (*sql.DB, bool) {
	// - Connect to SQL
	// Create the connection string to the root of the server.
	// Required if we need to create the Database first.
	connectionString := fmt.Sprintf("%s:%s@(%s:%d)/",
		*sqlUserName,
		*sqlPassword,
		*sqlHost,
		*sqlPort,
	)
	// Create the connecton. Go doesn't connect at this point, rahter
	// it waits till it needs to do something before making the connection.
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	// - Validate Connection
	// Make go create the connection to see if we are able to connect
	// to the server. Exit bad if we can't.
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	// - Check DB exists create if asked to;
	//   set created flag to not bother version Check
	// Flag to return to say that the Database is new and all
	// SQL files need to be run against it.
	newDB := false
	// Query the DB server to see if the DB exists.
	q := fmt.Sprintf("SHOW DATABASES LIKE '%s';", *sqldb)
	var holder string
	err = db.QueryRow(q).Scan(&holder)
	switch {
	case err == sql.ErrNoRows:
		// create the DB if asked to.
		if *createDB == true {
			log.Println("DB not found creating.")
			_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s;", *sqldb))
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
	_, err = db.Exec(fmt.Sprintf("USE %s;", *sqldb))
	if err != nil {
		log.Fatal(err)
	}

	// return back the db connection and the flag to tell if we created
	// the database.
	return db, newDB
}
