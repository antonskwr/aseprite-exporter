package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	RunCmdLiteral = "run"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 6 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CommandLine) runExporter() {

}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	runCmd := flag.NewFlagSet(RunCmdLiteral, flag.ExitOnError)

	asepritePath := runCmd.String("asepath", "", "Path to aseprite executable")
	sourceDir := runCmd.String("source", "", "Path to directory with aseprite projects")
	targetDir := runCmd.String("target", "", "Path to directory for project tree to be exported into")
	timeDbPath := runCmd.String("db", "", "DB path for keeping project's last modified time")

	if os.Args[1] == RunCmdLiteral {
		err := runCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	}

	if runCmd.Parsed() {
		if *asepritePath == "" {
			//err
			runCmd.Usage() // or cli.printUsage()
			os.Exit(1)
		}

		if *sourceDir == "" {
			//err
			os.Exit(1)
		}

		if *targetDir == "" {
			//err
			os.Exit(1)
		}

		if *timeDbPath == "" {
			//err
			os.Exit(1)
		}

		// proceed
		cli.runExporter()
	}
}
