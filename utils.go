package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strings"
)

func getSSLCert(ip string, timeout int, dialer *net.Dialer) (*x509.Certificate, error) {
	conn, err := tls.DialWithDialer(dialer, "tcp", ip, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	cert := conn.ConnectionState().PeerCertificates[0]
	return cert, nil
}

// IPsFromCIDR generates a slice of IP strings from the given CIDR
func IPsFromCIDR(cidr string, chanInput chan string, ports []string) error {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		for _, port := range ports {
			chanInput <- ip.String() + ":" + port
		}
	}

	return nil
}

// inc increments an IP address
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func extractNames(cert *x509.Certificate) []string {
	names := append([]string{cert.Subject.CommonName}, cert.DNSNames...)
	return names
}

func intakeFunction(chanInput chan string, ports []string, input string) {
	if _, err := os.Stat(input); err == nil {
		readFile, err := os.Open(input)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fileScanner := bufio.NewScanner(readFile)

		fileScanner.Split(bufio.ScanLines)

		for fileScanner.Scan() {
			line := fileScanner.Text()
			processInput(line, chanInput, ports)
		}
		readFile.Close()

	} else {
		for _, argItem := range strings.Split(input, ",") {
			processInput(argItem, chanInput, ports)
		}
	}
}

func isCIDR(value string) bool {
	return strings.Contains(value, `/`)
}

func isHostPort(value string) bool {
	return strings.Contains(value, `:`)
}

func processInput(argItem string, chanInput chan string, ports []string) {
	argItem = strings.TrimSpace(argItem)
	if isHostPort(argItem) {
		chanInput <- argItem
	} else if isCIDR(argItem) {
		err := IPsFromCIDR(argItem, chanInput, ports)
		if err != nil {
			panic("unable to parse CIDR" + argItem)
		}
	} else {
		for _, port := range ports {
			chanInput <- argItem + ":" + port
		}
	}
}
