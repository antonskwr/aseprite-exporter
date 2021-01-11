package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type DBEntry struct {
	Path    string
	ModTime string
}

var projectsToExport []string
var filenameFormat = "[t={tag}][f={frame}].png"
var DB []DBEntry
var tempDB []DBEntry

type Option = int32

const (
	WasModified Option = iota
	NotModified
	NotFound
	DBEmpty
)

const (
	scaling = "1"
)

var asepriteRunCmd = ""

func Handle(err error, message ...string) {
	msg := ""

	if len(message) == 1 {
		msg = message[0] + ":"
	}

	if err != nil {
		log.Panic(msg, err)
	}
}

// I form a slice of the above struct

// func parse() {
// 	for i := range myconfig {
// 		if myconfig[i].Key == "key1" {
// 			// Found!
// 			break
// 		}
// 	}
// }

func checkNotEmptyDB() bool {
	return len(DB) != 0
}

func loadDB(dbPath string) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("No DB Found")
		return
	}

	fileTextBuffer, err := ioutil.ReadFile(dbPath)
	Handle(err)

	sourceText := string(fileTextBuffer)
	lines := strings.Split(sourceText, "\n")
	lines = lines[:len(lines)-1]

	for i := range lines {
		entries := strings.Split(lines[i], "|")
		if len(entries) != 2 {
			log.Panicf(
				"DB entry at line:%v wrong num of entries | expected 2 | actual:%v\n",
				i+1,
				len(entries),
			)
		}

		dbEntry := DBEntry{entries[0], entries[1]}
		DB = append(DB, dbEntry)
	}
}

func createDBEntry(path string) DBEntry {
	file, err := os.Stat(path)

	if os.IsNotExist(err) {
		log.Panicf("createDBEntry file does't exist at path:%s\n", path)
	}

	modtime := file.ModTime().Format(time.UnixDate)

	return DBEntry{path, modtime}
}

func checkFileModified(entry DBEntry) Option {
	if !checkNotEmptyDB() {
		return DBEmpty
	}

	for i := range DB {
		if DB[i].Path == entry.Path {
			if DB[i].ModTime == entry.ModTime {
				return NotModified
			}
			// Found!
			break
		}
	}

	return NotFound
}

func tree(root string, exportPath string) error {
	err := filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
		if path.Ext(p) == ".aseprite" {

      var trimFlag = ""

			dbEntry := createDBEntry(p)
			tempDB = append(tempDB, dbEntry)

			switch checkFileModified(dbEntry) {
			case WasModified:
			case DBEmpty:
			case NotModified:
				fmt.Printf("was not modified: %s\n", p)
				return nil
			}

			pWithoutRoot := strings.TrimPrefix(p, path.Dir(root))
			pWithoutExtension := strings.TrimSuffix(pWithoutRoot, path.Ext(pWithoutRoot))
			exportDir := path.Join(exportPath, pWithoutExtension)
			fmt.Printf("export: %s\n", exportDir)

			err = os.MkdirAll(exportDir, os.ModePerm)

			Handle(err, "export error")

			filename := filepath.Base(exportDir)
			fullExpPath := fmt.Sprintf("%s/%s-%s", exportDir, filename, filenameFormat)

			if strings.HasSuffix(filename, "_s") {
				fullExpPath = fmt.Sprintf("%s/{layer}-%s", exportDir, filenameFormat)
				filename = strings.TrimSuffix(filename, "_s")
			}

			if strings.HasSuffix(filename, "_t") {
				trimFlag = "--trim"
			}

			out, err := exec.Command(asepriteRunCmd, "-b", p, "--scale", scaling, trimFlag, "--save-as", fullExpPath).Output()
			Handle(err, "aseprite error")

			fmt.Printf("%s", out)
		}

		return nil
	})

	return err
}

func updateDB(dbPath string) {
	file, err := os.Create(dbPath)
	Handle(err, "failed to update DB")
	defer file.Close()

	w := bufio.NewWriter(file)
	for i := range tempDB {
		line := fmt.Sprintf("%s|%s", tempDB[i].Path, tempDB[i].ModTime)
		fmt.Fprintln(w, line)
	}
	Handle(w.Flush(), "failed to update DB [flush]")
}

func main() {
	if len(os.Args) < 5 {
		log.Panic("missing arguments")
	}

	asepriteRunCmd = os.Args[1]
	asepriteProjectDir := os.Args[2]
	exportPath := os.Args[3]
	dbPath := os.Args[4]

	loadDB(dbPath)
	if !checkNotEmptyDB() {
		err := os.RemoveAll(exportPath)
		Handle(err)
	}

	err := tree(asepriteProjectDir, exportPath)
	if err != nil {
		log.Printf("tree %s: %v\n", os.Args[1], err)
	}

	updateDB(dbPath)
}
