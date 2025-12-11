package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func getHtml(urlStr string) (string, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get(urlStr)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	htmlBody, ioError := io.ReadAll(resp.Body)
	if ioError != nil {
		fmt.Println(ioError)
		return "", ioError
	}
	return string(htmlBody), nil
}

var visitedPages = make(map[string]bool)

// ---------------------------------------------------------------------
//
//	Поиск всех типов ресурсов (img, script, css, html‑страницы).
//
// ---------------------------------------------------------------------
func extractResourceTagsForSinglePage(url *url.URL, htmlBody string) []ResourceLink {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		fmt.Println(err)
	}

	var resourceLinks []ResourceLink

	for n := range doc.Descendants() {
		if n.Type != html.ElementNode {
			continue
		}

		switch n.DataAtom {
		case atom.Img, atom.Script:
			for _, a := range n.Attr {
				if a.Key == "src" {
					fmt.Println("image", a.Val)
					srcURL, _ := url.Parse(a.Val)
					resourceLinks = append(resourceLinks, *NewResourceLink(n.DataAtom, *srcURL))
				}
			}
		case atom.Style:
			if n.FirstChild != nil {
				fmt.Println("style", n.FirstChild.Data)
			}
		case atom.Link:
			// CSS‑файлы
			for _, a := range n.Attr {
				if a.Key == "href" && a.Val != "" &&
					(strings.HasSuffix(a.Val, ".css") || strings.Contains(strings.ToLower(getAttrValue(n, "rel")), "stylesheet")) {
					fmt.Println("stylesheet", a.Val)
					srcURL, _ := url.Parse(a.Val)
					resourceLinks = append(resourceLinks, *NewResourceLink(atom.Style, *srcURL))
				}
			}
		// -----------------------------------------------------------------
		//  Обрабатываем обычные ссылки <a href="…">
		// -----------------------------------------------------------------
		case atom.A:
			for _, a := range n.Attr {
				if a.Key == "href" && a.Val != "" {
					hrefValue := a.Val

					// отбрасываем «мусорные» ссылки
					if strings.HasPrefix(hrefValue, "#") ||
						strings.HasPrefix(strings.ToLower(hrefValue), "mailto:") ||
						strings.HasPrefix(strings.ToLower(hrefValue), "javascript:") ||
						hrefValue == "/" {
						continue
					}

					fmt.Println("anchor html page", hrefValue)
					srcURL, _ := url.Parse(a.Val)
					resourceLinks = append(resourceLinks, *NewResourceLink(n.DataAtom, *srcURL))
				}
			}
		}
	}
	return resourceLinks
}

// Вспомогательная функция для получения значения атрибута
func getAttrValue(n *html.Node, attrName string) string {
	for _, a := range n.Attr {
		if a.Key == attrName {
			return a.Val
		}
	}
	return ""
}

func CrawlWebSite(url url.URL) *WebSiteResourceGraph {
	return crawlPages(url, 1)
}

func crawlPages(url url.URL, recursionThreshold int) *WebSiteResourceGraph {
	mainPageHtmlBody, _ := getHtml(url.String())

	var mainPageResourceURLs = extractResourceTagsForSinglePage(&url, mainPageHtmlBody)
	var webSiteResourceGraph = NewWebSiteResourceGraph(url, []ResourceLink{}, mainPageHtmlBody)

	var pageLinks []ResourceLink
	var resourceLinksOnlyStatic []ResourceLink

	// Выбираем только ресурсы-ссылки
	for _, resourceLink := range mainPageResourceURLs {
		if resourceLink.LinkType() == atom.Link || resourceLink.LinkType() == atom.A {
			visited := visitedPages[resourceLink.String()]
			if visited {
				continue
			}
			pageLinks = append(pageLinks, resourceLink)
			visitedPages[resourceLink.String()] = true
		} else {
			resourceLinksOnlyStatic = append(resourceLinksOnlyStatic, resourceLink)
		}
	}

	webSiteResourceGraph.resourceLinks = resourceLinksOnlyStatic

	if recursionThreshold > 0 {
		for _, resourceLink := range pageLinks {
			webSiteResourceGraph.AddDescendant(*crawlPages(resourceLink.LinkURL(), recursionThreshold-1))
		}
	}

	return webSiteResourceGraph
}
