package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

// Setup the flags that we need.
// The default values are kept in here also. We use these for seeding the environment
// running config also.
var (
	sqlHost            = flag.String("sqlhost", "localhost", "Host of the SQL server. Use SQL_HOST in environment variable mode.")
	sqlPort            = flag.Int64("sqlport", 3306, "Port to connect to the SQL server on. Use SQL_PORT in environment variable mode.")
	sqlUserName        = flag.String("sqlusername", "root", "Username to connect to the SQL server with. Use SQL_USERNAME in environment variable mode.")
	sqlPassword        = flag.String("sqlpassword", "root", "Password to connect to the SQL server with. Use SQL_PASSWORD in environment variable mode.")
	sqldb              = flag.String("sqldb", "test", "Name of the DB that we should be connecting to. Use SQL_DB in environment variable mode.")
	sqlVersionTable    = flag.String("sqlversion-table", "version", "Name of the DB table that keeps version information. Use SQL_VERSION_TABLE in environment variable mode.")
	scriptsPath        = flag.String("scripts-path", "/data", "Directory containing the upgrade scripts to run. Use SCRIPTS_PATH in environment variable mode.")
	createDB           = flag.Bool("create-db", false, "Create the Database if it does not exist. Use CREATE_DB in environment variable mode.")
	version            = flag.Bool("v", false, "Prints out the version number")
	help               = flag.Bool("h", false, "Display Help message.")
	collectEnvironment = flag.Bool("use-environment-variables", false, "Should the application collect the configuration from environment variables.")
)

// RunConfig holds the values from either the flags passed in or the Environment variables.
type RunConfig struct {
	sqlHost         string
	sqlPort         int64
	sqlUserName     string
	sqlPassword     string
	sqldb           string
	sqlVersionTable string
	scriptsPath     string
	createDB        bool
	showVersion     bool
	sync.RWMutex
}

func (rc *RunConfig) scrapeEnvironment() {
	// Not happy with this and need to figure out a better way to do it.
	if value, ok := os.LookupEnv("SQL_HOST"); ok {
		rc.sqlHost = value
	}
	if value, ok := os.LookupEnv("SQL_PORT"); ok {
		portInt64, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Fatal("SQL_PORT is not a valid number.")
		}
		rc.sqlPort = portInt64
	}
	if value, ok := os.LookupEnv("SQL_USERNAME"); ok {
		rc.sqlUserName = value
	}
	if value, ok := os.LookupEnv("SQL_PASSWORD"); ok {
		rc.sqlPassword = value
	}
	if value, ok := os.LookupEnv("SQL_DB"); ok {
		rc.sqldb = value
	}
	if value, ok := os.LookupEnv("SQL_VERSION_TABLE"); ok {
		rc.sqlVersionTable = value
	}
	if value, ok := os.LookupEnv("SCRIPTS_PATH"); ok {
		rc.scriptsPath = value
	}
	if value, ok := os.LookupEnv("CREATE_DB"); ok {
		boolVar, err := strconv.ParseBool(value)
		if err != nil {
			log.Fatal("CREATE_DB is not a valid value. It must be one of the following [1,0,t,f,T,F,True,False,TRUE,FALSE,true,false]")
		}
		rc.createDB = boolVar
	}
}

// SQLConfig holds the functions that return config for the SQL server we
// should be talking to.
type SQLConfig interface {
	SQLHost() string
	SQLPort() int64
	SQLDB() string
	SQLUsername() string
	SQLPassword() string
	SQLVersionTable() string
	CreateDB() bool
}

// FileConfig These are the files that we need to parse
type FileConfig interface {
	ScriptsPath() string
}

func (rc *RunConfig) SQLHost() string {
	rc.RLock()
	defer rc.RUnlock()
	return rc.sqlHost
}
func (rc *RunConfig) SQLPort() int64 {
	rc.RLock()
	defer rc.RUnlock()
	return rc.sqlPort
}
func (rc *RunConfig) SQLDB() string {
	rc.RLock()
	defer rc.RUnlock()
	return rc.sqldb
}
func (rc *RunConfig) SQLUsername() string {
	rc.RLock()
	defer rc.RUnlock()
	return rc.sqlUserName
}
func (rc *RunConfig) SQLPassword() string {
	rc.RLock()
	defer rc.RUnlock()
	return rc.sqlPassword
}
func (rc *RunConfig) SQLVersionTable() string {
	rc.RLock()
	defer rc.RUnlock()
	return rc.sqlVersionTable
}
func (rc *RunConfig) CreateDB() bool {
	rc.RLock()
	defer rc.RUnlock()
	return rc.createDB
}
func (rc *RunConfig) ScriptsPath() string {
	rc.RLock()
	defer rc.RUnlock()
	return rc.scriptsPath
}
func (rc *RunConfig) ShowVersion() bool {
	rc.RLock()
	defer rc.RUnlock()
	return rc.showVersion
}

func newRunConfig() *RunConfig {
	return &RunConfig{
		sqlHost:         *sqlHost,
		sqlPort:         *sqlPort,
		sqldb:           *sqldb,
		sqlUserName:     *sqlUserName,
		sqlPassword:     *sqlPassword,
		sqlVersionTable: *sqlVersionTable,
		scriptsPath:     *scriptsPath,
		createDB:        *createDB,
		showVersion:     *version,
	}
}

// GetConfig will return the configuration to run with.
func GetConfig() *RunConfig {
	flag.Parse()

	// Check that if there are flags and if the help flag was set
	if *help == true || len(os.Args) < 2 {
		displayHelp()
		os.Exit(0)
	}

	conf := newRunConfig()

	if *collectEnvironment {
		conf.scrapeEnvironment()
	}

	return conf
}

// displayHelp Shows the help message
func displayHelp() {
	s := `This script is used to upgrade MySQL server schema.
It takes upgrade scripts that are numbered and applies them in order.
The numbers must appear at the begining of the file.
The can have gaps in them and can also have leading zeros if that
helps make things look better.

Examples: 1_something.sql, 002somethingelse.sql, 54_someOtherthing.sql

The database must have a table that will accomodate
the version numbers the database is on and also the names of the
files that have been run on the verson as well as the MD5 checksum. The table
by default is called version. If you need to retro fit this tool you can set
the name with the -sqlversion-table flag.

CREATE TABLE version(id int not NULL PRIMARY KEY, file TEXT(500), checksum TEXT(32));

If you would like to have the tool create the database and table then you should set the
-create-db flag. It will also create the version table.
`
	fmt.Println(s)
	flag.PrintDefaults()
}
