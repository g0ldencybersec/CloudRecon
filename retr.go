package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
)

type RetrArgs struct {
	NumResults      bool
	All             bool
	Help            bool
	Database        string
	QueryOrg        string
	QueryCommonName string
	QuerySAN        string
	QueryIP         string
}

type Certificate struct {
	IP           string `json:"ip"`
	Organization string `json:"organization"`
	CommonName   string `json:"common_name"`
	SAN          string `json:"san"`
}

func runCloudRetr(clArgs []string) {
	args := parseRetrCLI(clArgs)

	sqliteDatabase, _ := sql.Open("sqlite3", args.Database)

	if args.NumResults {
		fmt.Printf("Cert DB has results for %d IPs\n", getNumResults(sqliteDatabase))
		sqliteDatabase.Close()
		return
	}
	if args.All {
		fmt.Println(getAllResults(sqliteDatabase))
		sqliteDatabase.Close()
		return
	}

	if args.QueryIP != "NONE" {
		fmt.Println(queryByIP(sqliteDatabase, args.QueryIP))
		sqliteDatabase.Close()
		return
	}
	if args.QueryOrg != "NONE" {
		fmt.Println(queryByOrg(sqliteDatabase, args.QueryOrg))
		sqliteDatabase.Close()
		return
	}
	if args.QueryCommonName != "NONE" {
		fmt.Println(queryByCommonName(sqliteDatabase, args.QueryCommonName))
		sqliteDatabase.Close()
		return
	}
	if args.QuerySAN != "NONE" {
		fmt.Println(queryBySAN(sqliteDatabase, args.QuerySAN))
		sqliteDatabase.Close()
		return
	}

}

func parseRetrCLI(clArgs []string) RetrArgs {
	args := RetrArgs{}
	retrUsage := "retr [options]"

	retrCommand := flag.NewFlagSet("scrape", flag.ContinueOnError)
	retrCommand.BoolVar(&args.Help, "h", false, "print usage!")
	retrCommand.BoolVar(&args.NumResults, "num", false, "Return the Number of rows (results) in the DB (By IP)")
	retrCommand.BoolVar(&args.All, "all", false, "Return all the rows in the DB")
	retrCommand.StringVar(&args.Database, "db", "certificates.db", "String of the DB you want to connect to and save certs!")
	retrCommand.StringVar(&args.QueryOrg, "org", "NONE", "String to search for in Organization column, returns like-results")
	retrCommand.StringVar(&args.QueryCommonName, "cn", "NONE", "String to search for in common name column, returns like-results")
	retrCommand.StringVar(&args.QuerySAN, "san", "NONE", "String to search for in common name column, returns like-results")
	retrCommand.StringVar(&args.QueryIP, "ip", "NONE", "String to search for in IP column, returns like-results")

	retrCommand.Parse(clArgs)

	if args.Help {
		fmt.Println(retrUsage)
		retrCommand.PrintDefaults()
	}

	return args
}

func getNumResults(db *sql.DB) int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM certificates").Scan(&count)
	if err != nil {
		panic(err)
	}
	return count
}

func getAllResults(db *sql.DB) string {
	rows, err := db.Query("SELECT ip, organization, common_name, san FROM certificates")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var certs []Certificate
	for rows.Next() {
		var c Certificate
		err = rows.Scan(&c.IP, &c.Organization, &c.CommonName, &c.SAN)
		if err != nil {
			panic(err)
		}
		certs = append(certs, c)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	jsonData, err := json.MarshalIndent(certs, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonData)
}

func queryByOrg(db *sql.DB, searchTerm string) string {
	rows, err := db.Query(`SELECT ip, organization, common_name, san FROM certificates WHERE organization LIKE ?`, "%"+searchTerm+"%")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var certs []Certificate
	for rows.Next() {
		var c Certificate
		err = rows.Scan(&c.IP, &c.Organization, &c.CommonName, &c.SAN)
		if err != nil {
			panic(err)
		}
		certs = append(certs, c)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	jsonData, err := json.MarshalIndent(certs, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonData)
}

func queryByIP(db *sql.DB, searchTerm string) string {
	rows, err := db.Query(`SELECT ip, organization, common_name, san FROM certificates WHERE ip LIKE ?`, searchTerm+"%")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var certs []Certificate
	for rows.Next() {
		var c Certificate
		err = rows.Scan(&c.IP, &c.Organization, &c.CommonName, &c.SAN)
		if err != nil {
			panic(err)
		}
		certs = append(certs, c)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	jsonData, err := json.MarshalIndent(certs, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonData)
}

func queryByCommonName(db *sql.DB, searchTerm string) string {
	rows, err := db.Query(`SELECT ip, organization, common_name, san FROM certificates WHERE common_name LIKE ?`, "%"+searchTerm+"%")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var certs []Certificate
	for rows.Next() {
		var c Certificate
		err = rows.Scan(&c.IP, &c.Organization, &c.CommonName, &c.SAN)
		if err != nil {
			panic(err)
		}
		certs = append(certs, c)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	jsonData, err := json.MarshalIndent(certs, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonData)
}

func queryBySAN(db *sql.DB, searchTerm string) string {
	rows, err := db.Query(`SELECT ip, organization, common_name, san FROM certificates WHERE san LIKE ?`, "%"+searchTerm+"%")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var certs []Certificate
	for rows.Next() {
		var c Certificate
		err = rows.Scan(&c.IP, &c.Organization, &c.CommonName, &c.SAN)
		if err != nil {
			panic(err)
		}
		certs = append(certs, c)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	jsonData, err := json.MarshalIndent(certs, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonData)
}
