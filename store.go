package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type StoreArgs struct {
	Concurrency int
	Ports       []string
	Timeout     int
	PortList    string
	Help        bool
	Input       string
	Database    string
}

func runCloudStore(clArgs []string) {
	args := parseStoreCLI(clArgs)

	if _, err := os.Stat(args.Database); err == nil {
		fmt.Printf("Using database file %s\n", args.Database)
	} else {
		//Create DB file if it doesn't exist
		CreateDatabase(args.Database)
	}

	sqliteDatabase, _ := sql.Open("sqlite3", args.Database) // Open the created SQLite File
	createTable(sqliteDatabase)                             // Create Database Tables if needed

	dialer := &net.Dialer{
		Timeout: time.Duration(args.Timeout) * time.Second,
	}

	//Channel for input
	inputChannel := make(chan string)

	var inputwg sync.WaitGroup
	for i := 0; i < args.Concurrency; i++ {
		inputwg.Add(1)
		go func(db *sql.DB) {
			for ip := range inputChannel {
				cert, err := getSSLCert(ip, args.Timeout, dialer)
				if err != nil {
					continue
				} else {
					names := extractNames(cert)
					org := cert.Subject.Organization
					if len(org) > 0 {
						_, err = db.Exec("INSERT INTO certificates (ip, organization, common_name, san) VALUES (?, ?, ?, ?) ON CONFLICT(ip) DO UPDATE SET organization = excluded.organization, common_name = excluded.common_name, san = excluded.san", ip, org[0], names[0], strings.Join(names[1:], ","))
						if err != nil {
							panic(err)
						}
					} else {
						_, err = db.Exec("INSERT INTO certificates (ip, organization, common_name, san) VALUES (?, ?, ?, ?) ON CONFLICT(ip) DO UPDATE SET organization = excluded.organization, common_name = excluded.common_name, san = excluded.san", ip, "NONE", names[0], strings.Join(names[1:], ","))
						if err != nil {
							panic(err)
						}
					}
				}

			}
			inputwg.Done()
		}(sqliteDatabase)
	}

	intakeFunction(inputChannel, args.Ports, args.Input)
	close(inputChannel)
	inputwg.Wait()
	sqliteDatabase.Close()
}

func parseStoreCLI(clArgs []string) StoreArgs {
	args := StoreArgs{}
	storeUsage := "store [options] -i <IPs/CIDRs or File>"

	storeCommand := flag.NewFlagSet("scrape", flag.ContinueOnError)
	storeCommand.IntVar(&args.Concurrency, "c", 100, "How many goroutines running concurrently")
	storeCommand.StringVar(&args.PortList, "p", "443", "TLS ports to check for certificates")
	storeCommand.IntVar(&args.Timeout, "t", 4, "Timeout for TLS handshake")
	storeCommand.BoolVar(&args.Help, "h", false, "print usage!")
	storeCommand.StringVar(&args.Input, "i", "NONE", "Either IPs & CIDRs separated by commas, or a file with IPs/CIDRs on each line")
	storeCommand.StringVar(&args.Database, "db", "certificates.db", "String of the DB you want to connect to and save certs!")

	storeCommand.Parse(clArgs)

	if args.Input == "NONE" {
		fmt.Print("No input detected, please use the -i flag to add input!\n\n")
		fmt.Println(storeUsage)
		storeCommand.PrintDefaults()
		os.Exit(1)
	}

	if args.Help {
		fmt.Println(storeUsage)
		storeCommand.PrintDefaults()
	}

	args.Ports = strings.Split(args.PortList, ",")
	return args
}

func CreateDatabase(databaseName string) {
	os.Remove(databaseName) // I delete the file to avoid duplicated records.
	fmt.Println("Creating db...")
	file, err := os.Create(databaseName) // Create SQLite file
	if err != nil {
		panic(err.Error())
	}
	file.Close()
	fmt.Println("db created")
}

func createTable(db *sql.DB) {
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS certificates (ip TEXT PRIMARY KEY NOT NULL, organization TEXT, common_name TEXT, san TEXT)")
	if err != nil {
		panic(err)
	}
	statement.Exec()
}
