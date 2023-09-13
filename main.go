package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

var (
	targetURL  string
	wg         sync.WaitGroup
	mu         sync.Mutex
	visited    = make(map[string]bool)
	loginPages = make(map[string]bool)
)

func createBanner() {
	banner := `
	██╗      ██████╗  ██████╗ ██╗███╗   ██╗    ██████╗  █████╗  ██████╗ ███████╗     ██████╗██╗  ██╗███████╗ ██████╗██╗  ██╗
	██║     ██╔═══██╗██╔════╝ ██║████╗  ██║    ██╔══██╗██╔══██╗██╔════╝ ██╔════╝    ██╔════╝██║  ██║██╔════╝██╔════╝██║ ██╔╝
	██║     ██║   ██║██║  ███╗██║██╔██╗ ██║    ██████╔╝███████║██║  ███╗█████╗      ██║     ███████║█████╗  ██║     █████╔╝ 
	██║     ██║   ██║██║   ██║██║██║╚██╗██║    ██╔═══╝ ██╔══██║██║   ██║██╔══╝      ██║     ██╔══██║██╔══╝  ██║     ██╔═██╗ 
	███████╗╚██████╔╝╚██████╔╝██║██║ ╚████║    ██║     ██║  ██║╚██████╔╝███████╗    ╚██████╗██║  ██║███████╗╚██████╗██║  ██╗
	╚══════╝ ╚═════╝  ╚═════╝ ╚═╝╚═╝  ╚═══╝    ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚══════╝     ╚═════╝╚═╝  ╚═╝╚══════╝ ╚═════╝╚═╝  ╚═╝

	*************************************************************************************************************************



    	/ \__
    	(    @\___
    	/         O
      /   (_____/
     /_____/   U
 

	 
	*************************************************************************************************************************
	_________ _______  _______  _______  _               _________ _______  _        _______  _______ 
	\__   __/(  ____ )(  ____ \(  ___  )( (    /|        \__   __/(  ___  )( (    /|(  ____ \(  ____ )
	   ) (   | (    )|| (    \/| (   ) ||  \  ( |           ) (   | (   ) ||  \  ( || (    \/| (    )|
	   | |   | (____)|| (__    | (___) ||   \ | |           | |   | (___) ||   \ | || (__    | (____)|
	   | |   |     __)|  __)   |  ___  || (\ \) |           | |   |  ___  || (\ \) ||  __)   |     __)
	   | |   | (\ (   | (      | (   ) || | \   |           | |   | (   ) || | \   || (      | (\ (   
	___) (___| ) \ \__| )      | )   ( || )  \  |           | |   | )   ( || )  \  || (____/\| ) \ \__
	\_______/|/   \__/|/       |/     \||/    )_)           )_(   |/     \||/    )_)(_______/|/   \__/
																									  
                                                                                                      
                                                                                            
															 
																															
    `
	fmt.Println(banner)
}
func crawl(url string) {
	defer wg.Done()

	mu.Lock()
	visited[url] = true
	mu.Unlock()

	client := &http.Client{
		Timeout: time.Second * 50,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("HTTP Get hatası: %s - %s\n", url, err)
		return
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Printf("HTML parse hatası: %s - %s\n", url, err)
		return
	}

	if isLoginPage(doc) {
		mu.Lock()
		loginPages[url] = true
		mu.Unlock()
	}

	var visit func(n *html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					newURL := resolveURL(url, a.Val)
					mu.Lock()
					if !visited[newURL] {
						visited[newURL] = true
						wg.Add(1)
						go crawl(newURL)
						fmt.Printf("Tarandı: %s\n", newURL)
					}
					mu.Unlock()
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}

	fmt.Printf("Taranıyor: %s\n", url)
	visit(doc)
}

func showProgress() {
	fmt.Println("Tarama devam ediyor...")
}

func main() {
	createBanner()
	fmt.Print("Hedef URL: ")
	reader := bufio.NewReader(os.Stdin)
	targetURL, _ = reader.ReadString('\n')
	targetURL = strings.TrimSpace(targetURL)

	fmt.Println("Tarama başlıyor...")

	wg.Add(1)
	go crawl(targetURL)

	showProgress()

	wg.Wait()

	fmt.Println("Giriş Yapılan Sayfalar:")
	for url := range loginPages {
		fmt.Println(url)
	}
}

func isLoginPage(doc *html.Node) bool {
	return strings.Contains(htmlRender(doc), "login")
}

func resolveURL(baseURL, relativeURL string) string {
	return baseURL + relativeURL
}

func htmlRender(n *html.Node) string {
	var render func(*html.Node) string
	render = func(n *html.Node) string {
		var result string
		if n.Type == html.TextNode {
			result = n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			result += render(c)
		}
		return result
	}
	return render(n)
}
