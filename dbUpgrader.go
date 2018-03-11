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
// *5  make a MD5 checksum of the file
// *6. apply scripts in required order and write the version, file and MD5 sum
//    to the version table

package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/morfien101/mysql-schema-updater/config"
)

const (
	version = "0.2.0"
)

// main is what is called when we start execution
func main() {
	conf := config.GetConfig()
	// Gather files.
	if conf.ShowVersion() {
		fmt.Print(version)
		os.Exit(0)
	}

	files := getFiles(conf.ScriptsPath())
	// Check for duplication in file numbering
	validateOrder(files)
	// Execute the SQL in the files.
	runFiles(files, conf)
}
