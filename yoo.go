package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/panjf2000/ants"
)

const (
	yooUA = "yoo/0.0.1"
)

var (
	m           = flag.String("m", "GET", "")
	headers     = flag.String("h", "", "")
	body        = flag.String("d", "", "")
	contentType = flag.String("T", "text/html", "")

	c = flag.Int("c", 50, "")
	n = flag.Int("n", 200, "")
	t = flag.Int("t", 20, "")
)

var usage = `Usage: hey [options...] <url>

Options:
  -n  Number of requests to run. Default is 200.
  -c  Number of requests to run concurrently. Total number of requests cannot
      be smaller than the concurrency level. Default is 50.

  -m  HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
  -t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
  -d  HTTP request body.
  -T  Content-type, defaults to "text/html".
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit("")
	}

	num := *n
	conc := *c

	url := flag.Args()[0]
	method := strings.ToUpper(*m)

	header := make(http.Header)
	header.Set("Content-Type", *contentType)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		usageAndExit(err.Error())
	}
	ua := req.UserAgent()
	if ua != "" {
		ua = yooUA
	} else {
		ua += " " + yooUA
	}
	header.Set("User-Agent", ua)
	req.Header = header

	// 构造请求
	do_request := func() {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Fprintf(os.Stdout, "%s", string(body))
	}

	defer ants.Release()
	p, _ := ants.NewPoolWithFunc(conc, func(i interface{}) {
		do_request()
	})
	defer p.Release()

	for i := 0; i < num; i++ {
		_ = p.Invoke(int32(i))
	}
}

type headerSlice []string

func (h *headerSlice) String() string {
	return fmt.Sprintf("%s", *h)
}

func (h *headerSlice) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}
