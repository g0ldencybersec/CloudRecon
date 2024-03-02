package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type CertificateInfo struct {
	IP           string `json:"ip"`
	Organization string `json:"organization"`
	CommonName   string `json:"commonName"`
	SAN          string `json:"san"`
}

type ScrapeArgs struct {
	Concurrency int
	Ports       []string
	Timeout     int
	PortList    string
	Help        bool
	Input       string
	AllOutput   bool
	JSONOutput  bool
}

func runCloudScrape(clArgs []string) {
	args := parseScrapeCLI(clArgs)

	dialer := &net.Dialer{
		Timeout: time.Duration(args.Timeout) * time.Second,
	}

	// Channel for input
	inputChannel := make(chan string)

	var inputwg sync.WaitGroup
	for i := 0; i < args.Concurrency; i++ {
		inputwg.Add(1)
		go func(AllOutput bool) {
			for ip := range inputChannel {
				cert, err := getSSLCert(ip, args.Timeout, dialer)
				if err != nil {
					if AllOutput {
						fmt.Printf("Failed to get SSL certificate from %s\n", ip)
					}
					continue
				} else {
					names := extractNames(cert)
					org := cert.Subject.Organization

					certInfo := CertificateInfo{
						IP:           ip,
						Organization: getOrganization(org),
						CommonName:   names[0],
						SAN:          joinNonEmpty(", ", names[1:]),
					}

					if args.JSONOutput {
						outputJSON, _ := json.Marshal(certInfo)
						fmt.Println(string(outputJSON))
					} else {
						fmt.Printf("Got SSL certificate from %s: [%s]\n", ip, strings.Join(names, ", "))
					}
				}
			}
			inputwg.Done()
		}(args.AllOutput)
	}

	intakeFunction(inputChannel, args.Ports, args.Input)
	close(inputChannel)
	inputwg.Wait()
}


func getOrganization(org []string) string {
	if len(org) > 0 {
			return org[0]
	}
	return "NONE"
}

func joinNonEmpty(sep string, elements []string) string {
	var result string
	for _, element := range elements {
			if element != "" {
					if result != "" {
							result += sep
					}
					result += element
			}
	}
	return result
}

func parseScrapeCLI(clArgs []string) ScrapeArgs {
	args := ScrapeArgs{}
	scrapeUsage := "scrape [options] -i <IPs/CIDRs or File>"

	scrapeCommand := flag.NewFlagSet("scrape", flag.ContinueOnError)
	scrapeCommand.IntVar(&args.Concurrency, "c", 100, "How many goroutines running concurrently")
	scrapeCommand.StringVar(&args.PortList, "p", "443", "TLS ports to check for certificates")
	scrapeCommand.IntVar(&args.Timeout, "t", 4, "Timeout for TLS handshake")
	scrapeCommand.BoolVar(&args.Help, "h", false, "print usage!")
	scrapeCommand.StringVar(&args.Input, "i", "NONE", "Either IPs & CIDRs separated by commas, or a file with IPs/CIDRs on each line")
	scrapeCommand.BoolVar(&args.AllOutput, "a", false, "Add this flag if you want to see all output including failures")
	scrapeCommand.BoolVar(&args.JSONOutput, "j", false, "Generate JSON output")

	scrapeCommand.Parse(clArgs)

	if args.Input == "NONE" {
		fmt.Print("No input detected, please use the -i flag to add input!\n\n")
		fmt.Println(scrapeUsage)
		scrapeCommand.PrintDefaults()
		os.Exit(1)
	}

	if args.Help {
		fmt.Println(scrapeUsage)
		scrapeCommand.PrintDefaults()
	}

	args.Ports = strings.Split(args.PortList, ",")
	return args
}
