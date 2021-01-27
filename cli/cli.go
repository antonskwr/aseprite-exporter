package cli

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/antonskwr/aseprite-exporter/exporter"
)

const (
	runCmdLiteral    = "export"
	asepritePathFlag = "execpath"
	sourceDirFlag    = "source"
	targetDirFlag    = "target"
	timeDbPathFlag   = "db"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Printf(
		"  %s -%s EXECUTABLE_PATH -%s SOURCE_DIR -%s TARGET_DIR -%s MODIFIED_TIME_DB\n",
		runCmdLiteral,
		asepritePathFlag,
		sourceDirFlag,
		targetDirFlag,
		timeDbPathFlag,
	)
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 6 {
		cli.printUsage()
		os.Exit(0)
	}
}

func (cli *CommandLine) runExporter() {

}

func (cli *CommandLine) Run() {
	runCmd := flag.NewFlagSet(runCmdLiteral, flag.ExitOnError)

	asepritePath := runCmd.String(asepritePathFlag, "", "Path to aseprite executable")
	sourceDir := runCmd.String(sourceDirFlag, "", "Path to directory with aseprite projects")
	targetDir := runCmd.String(targetDirFlag, "", "Path to directory for project tree to be exported into")
	timeDbPath := runCmd.String(timeDbPathFlag, "", "DB path for keeping project's last modified time")

	if len(os.Args) < 6 {
		cli.printUsage()
		runCmd.Usage()
		os.Exit(0)
	}

	if os.Args[1] == runCmdLiteral {
		err := runCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	} else {
		runCmd.Usage()
		os.Exit(0)
	}

	if runCmd.Parsed() {
		if *asepritePath == "" {
			runCmd.Usage()
			os.Exit(0)
		}

		if *sourceDir == "" {
			runCmd.Usage()
			os.Exit(0)
		}

		if *targetDir == "" {
			runCmd.Usage()
			os.Exit(0)
		}

		if *timeDbPath == "" {
			runCmd.Usage()
			os.Exit(0)
		}

		// proceed
		exporter.Run(*asepritePath, *sourceDir, *targetDir, *timeDbPath)
	}
}
