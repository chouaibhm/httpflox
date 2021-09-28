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
        //"regexp"
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

        timeout := time.Duration(to * 1000000) //- 0 for more speed

        var tr = &http.Transport{
                MaxIdleConns:        30,
                IdleConnTimeout:     time.Second,
                MaxIdleConnsPerHost: -1,
                DisableKeepAlives:   true,
                TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
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
                portlist := []string{"80","81","300","443","591","593","832","981","1010","1311","2082","2087","2095","2096","2480","3000","3128","3333","4444","4243","4567","4711","4712","4993","5000","5104","5108","5800","6543","7000","7396","7474","8000","8001","8008","8014","8042","8069","8080","8081","8088","8090","8091","8118","8123","8172","8222","8243","8280","8281","8333","8443","8500","8834","8880","8888","8983","9000","9043","9060","9080","9090","9091","9200","9443","9800","9981","12443","16080","18091","18092","20720","28017"}

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
        req.Header.Add("Connection", "close")
        req.Close = true

        resp, _ := client.Do(req)

        // Get page status
        if resp != nil {
                status := strconv.Itoa(resp.StatusCode)
                doc, _ := goquery.NewDocumentFromReader(resp.Body)
                if doc != nil {
                        //parse page for title
                        title := doc.Find("title").Contents().Text()
                        b := "[" + status + "]" + "  " + "[ " + strings.Replace(title, "\n","",-1) + " ]"
                        defer resp.Body.Close()

                        return b

                }
        }

        return ""

}
