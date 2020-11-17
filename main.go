package main

import (
        "bufio"
        "crypto/tls"
        "flag"
        "fmt"
        "github.com/PuerkitoBio/goquery"
        "io"
        "io/ioutil"
        "log"
        "net"
        "net/http"
        "os"
        "strconv"
        "strings"
        "sync"
        "time"
)

func main() {

        // concurrency level
        var concurrency int
        flag.IntVar(&concurrency, "c", 100, "set the concurrency")

        // timeout flag for giving up for the host:port
        var to int
        flag.IntVar(&to, "t", 3000, "timeout (milliseconds) default 3000")
        //silent mode

        //Silent mode

        silent := flag.Bool("silent", false, "run in silent mode ")
        // HTTP method to use
        var method string
        flag.StringVar(&method, "method", "HEAD", "HTTP method to use")

        flag.Parse()
        //TLS stuff

        timeout := time.Duration(to * 1000000)

        var tr = &http.Transport{
                MaxIdleConns:    30,
                IdleConnTimeout: time.Second,
                //MaxIdleConnsPerHost: -1,
                DisableKeepAlives: true,
                TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
                DialContext: (&net.Dialer{
                        Timeout:   timeout,
                        KeepAlive: time.Second,
                }).DialContext,
        }

        re := func(req *http.Request, via []*http.Request) error {
                return http.ErrUseLastResponse
        }

        client := &http.Client{
                Transport:     tr,
                CheckRedirect: re,
                Timeout:       timeout,
        }

        // make channels
        httpsURLs := make(chan string)
        httpURLs := make(chan string)
        output := make(chan string)

        // HTTPS workers
        var httpsWG sync.WaitGroup
        for i := 0; i < concurrency/2; i++ {
                httpsWG.Add(1)

                go func() {
                        for url := range httpsURLs {

                                // always try HTTPS first
                                withProto := "https://" + url
                                if isListening(client, withProto, method) {
                                        output <- withProto

                                }

                                httpURLs <- url
                        }

                        httpsWG.Done()
                }()
        }

        // HTTP workers
        var httpWG sync.WaitGroup
        for i := 0; i < concurrency/2; i++ {
                httpWG.Add(1)

                go func() {
                        for url := range httpURLs {
                                withProto := "http://" + url
                                if isListening(client, withProto, method) {
                                        output <- withProto
                                        continue
                                }
                        }

                        httpWG.Done()
                }()
        }

        // Close the httpURLs channel when the HTTPS workers are done
        go func() {
                httpsWG.Wait()
                close(httpURLs)
        }()

        // Output worker
        var outputWG sync.WaitGroup
        outputWG.Add(1)
        go func() {
                for out := range output {
                        if *silent == false {
                                fmt.Println(out + " " + status_title(client, out)) //if silent off
                        } else {
                                fmt.Println(out) //if silent on
                        }

                }
                outputWG.Done()
        }()

        // Close the output channel when the HTTP workers are done
        go func() {
                httpWG.Wait()
                close(output)
        }()

        // accept domains on stdin
        sc := bufio.NewScanner(os.Stdin)
        for sc.Scan() {
                domain := strings.ToLower(sc.Text())
                // Adding port list
                httpsURLs <- domain
                portlist := []string{"8080", "8081", "8082", "8443", "8181", "8081", "8888", "9200", "9090","7001"}

                for _, port := range portlist {
                        httpsURLs <- fmt.Sprintf("%s:%s", domain, port)
                }

        }

        // close channel
        close(httpsURLs)

        // check there were no errors reading stdin (unlikely)
        if err := sc.Err(); err != nil {
                fmt.Fprintf(os.Stderr, "failed to read input: %s\n", err)
        }

        // Wait until the output waitgroup is done
        outputWG.Wait()
}

func isListening(client *http.Client, url, method string) bool {

        req, err := http.NewRequest(method, url, nil)
        if err != nil {
                return false
        }

        req.Header.Add("Connection", "close")
        req.Close = true

        resp, err := client.Do(req)
        if resp != nil {
                io.Copy(ioutil.Discard, resp.Body)
                resp.Body.Close()
        }

        if err != nil {
                return false
        }

        return true
}

func status_title(client *http.Client, url string) string {

        req, err := http.NewRequest("GET", url, nil)
        if err != nil {
                log.Fatal(err)
        }

        resp, err := client.Do(req)

        //resp, err := http.Get(url)

        if err != nil {
                log.Fatal(err)
        }
        // Get page status
        status := strconv.Itoa(resp.StatusCode)
        //parse page for title
        doc, err := goquery.NewDocumentFromReader(resp.Body)
        if err != nil {
                log.Fatal(err)
        }

        title := doc.Find("title").Contents().Text()
        b := "[" + status + "]" + "  " + "[ " + string(title) + " ]"
        return b

}
