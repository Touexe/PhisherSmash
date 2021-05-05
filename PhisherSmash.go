package main

import (
	"bufio"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	url string = "https://nickmcgough.com/log.php"
	//url    string = "https://updatemessenger-826329.ingress-daribow.easywp.com/meaw/a7a.php"
	method string = "POST"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}

type Resp struct {
	*http.Response
	err error
}

type credential struct {
	username string
	password string
}

// sempaphore
var sem = make(chan struct{}, 200)

func RandomString(options []string) string {
	rand.Seed(time.Now().Unix())
	randNum := rand.Int() % len(options)
	return options[randNum]
}
func makeResp(credential credential, n *sync.WaitGroup, rc chan Resp, requestDone chan credential) {

	sem <- struct{}{}        // acquire
	defer func() { <-sem }() // release
	defer n.Done()
	req, _ := http.NewRequest(method, url, nil)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req.Header.Set("User-Agent", RandomString(userAgents))
	req.Header.Set("username", credential.username)
	req.Header.Set("password", credential.password)
	res, err := client.Do(req)
	r := Resp{res, err}
	rc <- r
	requestDone <- credential

}

func main() {
	runtime.GOMAXPROCS(3)
	var n sync.WaitGroup
	rc := make(chan Resp)

	file, err := os.Open("100k.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	timeStart := time.Now()
	requestDone := make(chan credential)

	go func() {
		reqs := 0
		for range requestDone {
			reqs++
			log.Printf("request %d done at %v", reqs, time.Since(timeStart))
		}
	}()

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ":")
		cred := credential{line[0], line[1]}
		n.Add(1)
		go makeResp(cred, &n, rc, requestDone)
	}

	go func() {
		n.Wait()
		close(rc)
	}()

	for result := range rc {
		statuscode := result.StatusCode
		if statuscode != 302 {
			log.Printf("Problem! HTTP STATUS CODE : %d", statuscode)

		}

	}

}
