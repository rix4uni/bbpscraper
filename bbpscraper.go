package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rix4uni/bbpscraper/banner"
)

var (
	pathFile     = flag.String("path", "", "File with list of paths to append")
	stopCount    = flag.Int("stop", 1, "Stop after N successful matches per domain (use 0 to check all paths)")
	summaryPath  = flag.String("summary", "", "File to write path match summary")
	timeoutSec   = flag.Int("timeout", 15, "Timeout per request in seconds")
	parallel     = flag.Int("parallel", 10, "Number of concurrent domain scans")
	mmc = flag.Int("mmc", 2, "Minimum number of regex matches to consider a valid result")
	silent = flag.Bool("silent", false, "silent mode.")
	version = flag.Bool("version", false, "Print the version of the tool and exit.")
	verbose = flag.Bool("verbose", false, "Enable verbose output for debugging")

	regexStr = `(?i)scope|Eligible Targets|reward|bounty|monetary|compensation|we offer a monetary|We offer reward|monetary reward|eligible for a reward|we award a bounty|We offer monetary rewards|security@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|bugbounty@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|bugreport@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|vulnerability@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|security-team@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|responsible-disclosure@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|infosec@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|security-alert@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|secure@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|security-alerts@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|security-notification@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|secure-report@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|security-incident@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|security-response@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|cybersecurity@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|reportabug@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|security-research@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}|security-reports@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,3}`
)

func readPaths(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var paths []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		if path != "" {
			paths = append(paths, path)
		}
	}
	return paths, scanner.Err()
}

func fetchAndMatch(url string, re *regexp.Regexp) ([]string, error) {
	client := &http.Client{Timeout: time.Duration(*timeoutSec) * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes)

	matches := re.FindAllString(body, -1)
	unique := make(map[string]bool)
	for _, match := range matches {
		normalized := strings.ToLower(strings.TrimSpace(match))
		unique[normalized] = true
	}

	var results []string
	for k := range unique {
		results = append(results, k)
	}
	return results, nil
}

func scanDomain(base string, paths []string, re *regexp.Regexp, summaryMu *sync.Mutex, summaryCounter map[string]int, wg *sync.WaitGroup) {
	defer wg.Done()

	matchCounter := 0

	for _, path := range paths {
		fullURL := strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")

		if *verbose {
			fmt.Printf("[DEBUG] Scanning %s\n", fullURL)
		}

		matches, err := fetchAndMatch(fullURL, re)
		if err != nil {
			continue
		}

		if len(matches) >= *mmc {
			formatted := "[\"" + strings.Join(matches, "\", \"") + "\"]"
			if *verbose {
				fmt.Printf("[FOUND] %s %s\n", fullURL, formatted)
			} else {
				fmt.Printf("%s %s\n", fullURL, formatted)
			}
			matchCounter++

			u, err := url.Parse(fullURL)
			if err == nil {
				summaryMu.Lock()
				summaryCounter[u.EscapedPath()]++
				summaryMu.Unlock()
			}

			if *stopCount > 0 && matchCounter >= *stopCount {
				break
			}
		}
	}
}

func main() {
	flag.Parse()

	if *version {
		banner.PrintBanner()
		banner.PrintVersion()
		return
	}

	if !*silent {
		banner.PrintBanner()
	}

	if *pathFile == "" {
		fmt.Println("Usage: bbpscraper -path bbp-paths.txt [-summary summary.txt] [-stop N] [-parallel N]")
		return
	}

	paths, err := readPaths(*pathFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading path file: %v\n", err)
		os.Exit(1)
	}

	re := regexp.MustCompile(regexStr)
	summaryCounter := make(map[string]int)
	var summaryMu sync.Mutex

	domains := []string{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain != "" {
			domains = append(domains, domain)
		}
	}

	sem := make(chan struct{}, *parallel)
	var wg sync.WaitGroup

	for _, domain := range domains {
		sem <- struct{}{}
		wg.Add(1)
		go func(d string) {
			defer func() { <-sem }()
			scanDomain(d, paths, re, &summaryMu, summaryCounter, &wg)
		}(domain)
	}

	wg.Wait()

	if *summaryPath != "" {
		f, err := os.Create(*summaryPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing summary file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		for path, count := range summaryCounter {
			fmt.Fprintf(f, "%s [count: %d]\n", path, count)
		}
	}
}
