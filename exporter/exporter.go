package exporter

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

type Option = int32

const (
	WasModified Option = iota
	NotModified
	NotFound
	DBEmpty
)

func handleErr(err error, message ...string) {
	if err != nil {
		if len(message) > 0 {
			err = fmt.Errorf("[%s] -- %w --", message[0], err)
		}
		log.Fatal(err)
	}
}

func checkNotEmptyDB(db *[]DBEntry) bool {
	return len(*db) != 0
}

func loadDB(dbPath string) []DBEntry {
	var DB []DBEntry

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("No DB Found")
		return DB
	}

	fileTextBuffer, err := ioutil.ReadFile(dbPath)
	handleErr(err)

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

	return DB
}

func createDBEntry(path string) DBEntry {
	file, err := os.Stat(path)

	if os.IsNotExist(err) {
		log.Panicf("createDBEntry file does't exist at path:%s\n", path)
	}

	modtime := file.ModTime().Format(time.UnixDate)

	return DBEntry{path, modtime}
}

func checkFileModified(entry DBEntry, db *[]DBEntry) Option {
	if !checkNotEmptyDB(db) {
		return DBEmpty
	}

	for i := range *db {
		if (*db)[i].Path == entry.Path {
			if (*db)[i].ModTime == entry.ModTime {
				return NotModified
			}
			// Found!
			break
		}
	}

	return NotFound
}

func tree(root string, exportPath string, db *[]DBEntry, expFunc ExportFunc) ([]DBEntry, error) {
	var tempDB []DBEntry

	err := filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
		if path.Ext(p) == ".aseprite" {

			var trimFlag = ""
			var scaling = "1"
			var filenameFormat = "[t={tag}][f={frame}].png"

			dbEntry := createDBEntry(p)
			tempDB = append(tempDB, dbEntry)

			switch checkFileModified(dbEntry, db) {
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

			handleErr(err, "export error")

			filename := filepath.Base(exportDir)
			exportedFileName := fmt.Sprintf("%s-%s", exportDir, filenameFormat)

			switch {
			case strings.HasSuffix(filename, "_t_s"):
				trimFlag = "--trim"
				exportedFileName = fmt.Sprintf("{layer}-%s", filenameFormat)
			case strings.HasSuffix(filename, "_s"):
				exportedFileName = fmt.Sprintf("{layer}-%s", filenameFormat)
			case strings.HasSuffix(filename, "_t"):
				trimFlag = "--trim"
				filename = strings.TrimSuffix(filename, "_t")
				exportedFileName = fmt.Sprintf("%s-%s", filename, filenameFormat)
			}

			fullExpPath := fmt.Sprintf("%s/%s", exportDir, exportedFileName)

			out, err := expFunc(p, scaling, trimFlag, fullExpPath)
			handleErr(err, "aseprite error")
			fmt.Printf("%s", out)
		}

		return nil
	})

	return tempDB, err
}

func updateDB(dbPath string, newDB *[]DBEntry) {
	file, err := os.Create(dbPath)
	handleErr(err, "failed to update DB")
	defer file.Close()

	w := bufio.NewWriter(file)
	for i := range *newDB {
		line := fmt.Sprintf("%s|%s", (*newDB)[i].Path, (*newDB)[i].ModTime)
		fmt.Fprintln(w, line)
	}
	handleErr(w.Flush(), "failed to update DB [flush]")
}

type ExportFunc func(filePath, scaling, trimFlag, exportPath string) ([]byte, error)

func newExportFunc(aseRunCmd string) ExportFunc {
	return func(filePath, scaling, trimFlag, exportPath string) ([]byte, error) {
		out, err := exec.Command(
			aseRunCmd,
			"-b", filePath,
			"--scale", scaling,
			trimFlag,
			"--save-as",
			exportPath,
		).Output()

		return out, err
	}
}

func checkDirExists(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: no such file or directory", path)
		}
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	return nil
}

func promptOverwriteWarning(path string) {
	fmt.Printf("Will overwrite \"%s\" directory, continue? [Y/N]\n", path)
	cliReader := bufio.NewReader(os.Stdin)
	text, err := cliReader.ReadString('\n')
	handleErr(err)

	text = strings.TrimSuffix(text, "\n")

	switch true {
	case (text == "Y" || text == "y"):
		return
	case (text == "N" || text == "n"):
		os.Exit(0)
	default:
		promptOverwriteWarning(path)
	}
}

func Run(aseRunCmd, sourceDir, targetDir, dbPath string, mutePrompt bool) {
	handleErr(checkDirExists(sourceDir))
	handleErr(checkDirExists(targetDir))

	db := loadDB(dbPath)
	if !checkNotEmptyDB(&db) {
		if !mutePrompt {
			promptOverwriteWarning(targetDir)
		}
		err := os.RemoveAll(targetDir) // this is dangerous
		handleErr(err)
	}

	exportFunc := newExportFunc(aseRunCmd)
	newDB, err := tree(sourceDir, targetDir, &db, exportFunc)
	if err != nil {
		msg := fmt.Sprintf("Failed to tree at %s", sourceDir)
		handleErr(err, msg)
	}

	updateDB(dbPath, &newDB)
}
