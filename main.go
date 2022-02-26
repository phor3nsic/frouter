package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func reqUrl(url string) string {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	statusCode := "0"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return statusCode
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:96.0) Gecko/20100101 Firefox/96.0")

	resp, err := client.Do(req)

	if err != nil {
		return statusCode
	}
	statusCode = strconv.Itoa(resp.StatusCode)
	return statusCode

}

func check_error(url string) bool {
	uri := "eW91X2lzX2Ffc3VwZXJIQWNrCg=="
	statusCode := reqUrl(url + uri)
	if statusCode == "404" {
		return true
	} else {
		return false

	}
}

func check_success(url string, uri string, wg *sync.WaitGroup, semaphore <-chan bool) {
	defer wg.Done()
	statusCode := reqUrl(url + uri)
	if statusCode[0:1] == "3" {
		fmt.Println(url + uri)
	}
	<-semaphore

}

func parse_url(url string) string {
	if url[len(url)-1:] != "/" {
		return url + "/"
	} else {
		return url
	}
}

func readWordlist(path string, url string, threads int) {
	wg := new(sync.WaitGroup)
	resultChan := make(chan string)
	file, _ := os.Open(path)
	scanner := bufio.NewScanner(file)
	semaphore := make(chan bool, threads)
	lineNo := 0

	for scanner.Scan() {
		lineNo = lineNo + 1
		semaphore <- true
		wg.Add(1)
		go check_success(url, scanner.Text(), wg, semaphore)

	}
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- true
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("error: ", err)
		close(resultChan)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		fmt.Println(result)
	}

}

func readStdin(wordlist string, threads int) {
	var line string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line = scanner.Text()
		url := parse_url(line)
		fmt.Println(url)

		if check_error(url) == true {
			readWordlist(wordlist, url, threads)
		}
	}

}

var (
	wordlist string
	threads  int
	proxy    string
)

func main() {
	flag.StringVar(&wordlist, "w", "", "Wordlist to usage")
	flag.IntVar(&threads, "t", runtime.GOMAXPROCS(0), "Threads to job")
	flag.StringVar(&proxy, "p", "", "Proxy to debug")
	flag.Parse()

	if proxy != "" {
		os.Setenv("HTTP_PROXY", proxy)
	}

	readStdin(wordlist, threads)

}
