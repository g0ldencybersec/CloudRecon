package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"time"
)

type Result struct {
	IP          string
	Hit         bool
	Error       error
	Certificate *CertificateInfo
}

type Worker struct {
	dialer  *net.Dialer
	input   <-chan string
	results chan<- Result
}

type WorkerPool struct {
	workers []*Worker
	input   chan string
	results chan Result
	dialer  *net.Dialer
}

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

func NewWorker(dialer *net.Dialer, input <-chan string, results chan<- Result) *Worker {
	return &Worker{
		dialer:  dialer,
		input:   input,
		results: results,
	}
}

func (w *Worker) run() {
	for ip := range w.input {
		cert, err := getSSLCert(ip, w.dialer)
		if err != nil {
			w.results <- Result{IP: ip, Error: err}
			continue
		}

		names := extractNames(cert)
		org := cert.Subject.Organization

		certInfo := CertificateInfo{
			IP:           ip,
			Organization: getOrganization(org),
			CommonName:   names[0],
			SAN:          joinNonEmpty(", ", names[1:]),
		}

		w.results <- Result{IP: ip, Hit: true, Certificate: &certInfo}
	}
}

func NewWorkerPool(size int, dialer *net.Dialer, input chan string, results chan Result) *WorkerPool {
	wp := &WorkerPool{
		workers: make([]*Worker, size),
		input:   input,
		results: results,
		dialer:  dialer,
	}
	for i := range wp.workers {
		wp.workers[i] = NewWorker(wp.dialer, wp.input, wp.results)
	}
	return wp
}

func (wp *WorkerPool) Start() {
	for _, worker := range wp.workers {
		go worker.run()
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.input)
}

func runCloudScrape(clArgs []string) {
	args := parseScrapeCLI(clArgs)

	dialer := &net.Dialer{
		Timeout: time.Duration(args.Timeout) * time.Second,
	}

	inputChannel := make(chan string)
	resultChannel := make(chan Result)

	workerPool := NewWorkerPool(args.Concurrency, dialer, inputChannel, resultChannel)
	workerPool.Start()

	go intakeFunction(inputChannel, args.Ports, args.Input)

	defer func() {
		workerPool.Stop()
		close(resultChannel)
	}()

	for result := range resultChannel {
		if result.Error != nil {
			fmt.Printf("Failed to get SSL certificate from %s: %v\n", result.IP, result.Error)
		} else if result.Hit {
			if args.JSONOutput {
				outputJSON, _ := json.Marshal(result.Certificate)
				fmt.Println(string(outputJSON))
			} else {
				fmt.Printf("Got SSL certificate from %s: [%s]\n", result.IP, result.Certificate.CommonName)
			}
		} else if args.AllOutput {
			fmt.Printf("No SSL certificate found for %s\n", result.IP)
		}
	}
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
