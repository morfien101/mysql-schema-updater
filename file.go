package main

// File holds data about each file we have
import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/morfien101/mysql-schema-updater/md5er"
)

type File struct {
	Index int
	Path  string
}

// String allows the type to be output as a sting native value.
func (f File) String() string {
	return fmt.Sprintf("%d - %s", f.Index, f.Path)
}

func (f File) md5Hash() string {
	md5hash, err := md5er.HashFile(f.Path)
	if err != nil {
		log.Fatalf("Failed to hash file %s. Error: %s", f.Path, err)
	}
	return md5hash
}

// Implement the sorter and functions
type byIndex []File

func (a byIndex) Len() int           { return len(a) }
func (a byIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byIndex) Less(i, j int) bool { return a[i].Index < a[j].Index }

// getFiles gets the file names from disk, sorts them
// and returns a slice of type Files. This will have
// the file name and its digested id.
func getFiles(path string) []File {
	filesSlice := gatherFiles(path)
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
		if len(sqlfiles) == 0 {
			log.Fatal("Did not file and files to process")
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
