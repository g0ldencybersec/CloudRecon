package main

import (
	"flag"
	"fmt"
	"os"
	"path"
)

var mainUsage = "scrape|store|retr [options]"

func commandUsage(msg string, cmdFlagSet *flag.FlagSet) {
	fmt.Printf("Usage: %s %s\n\n", path.Base(os.Args[0]), msg)
	cmdFlagSet.PrintDefaults()

	if msg == mainUsage {
		fmt.Printf("\nSubcommands: \n\n")
		fmt.Printf("\t%-11s - Scrape given IPs and output CNs & SANs to stdout\n", "cloudrecon scrape")
		fmt.Printf("\t%-11s - Scrape and collect Orgs,CNs,SANs in local db file\n", "cloudrecon store")
		fmt.Printf("\t%-11s - Query local DB file for results\n", "cloudrecon retr")
	}
}

func main() {

	mainFlagSet := flag.NewFlagSet("amass", flag.ContinueOnError)
	var help bool
	mainFlagSet.BoolVar(&help, "h", false, "Show the program usage message")

	if len(os.Args) < 2 {
		commandUsage(mainUsage, mainFlagSet)
		return
	}
	if err := mainFlagSet.Parse(os.Args[1:]); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	if help {
		commandUsage(mainUsage, mainFlagSet)
		return
	}

	switch os.Args[1] {
	case "scrape":
		runCloudScrape(os.Args[2:])
	case "store":
		runCloudStore(os.Args[2:])
	case "retr":
		runCloudRetr(os.Args[2:])
	case "help":
		commandUsage(mainUsage, mainFlagSet)
		return
	default:
		commandUsage(mainUsage, mainFlagSet)
		os.Exit(1)
	}
}
