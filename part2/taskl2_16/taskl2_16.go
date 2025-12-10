package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"taskl2_16/ResourceTypes"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type AnyResource = ResourceTypes.AnyResource

func getHtml(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
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

func extractResourceTags(url *url.URL, htmlBody string) []AnyResource {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		fmt.Println(err)
	}

	var resourceTags []AnyResource

	for n := range doc.Descendants() {
		if n.Type != html.ElementNode {
			continue
		}

		switch n.DataAtom {
		case atom.Img:
			for _, a := range n.Attr {
				if a.Key == "src" {
					srcValue := a.Val
					fmt.Println("image", srcValue)
					resourceTags = append(
						resourceTags,
						ResourceTypes.ImageResource{SourceUrlString: url.String(), PathString: srcValue})
				}
			}
		case atom.Style:
			if n.FirstChild != nil {
				fmt.Println("style", n.FirstChild.Data)
			}
		case atom.Script:
			//for _, a := range n.Attr {
			//	if a.Key == "src" {
			//		srcValue := a.Val
			//		fmt.Println("script", srcValue)
			//		resourceTags = append(
			//			resourceTags,
			//			ResourceTypes.ScriptResource{SourceUrlString: url.String(), PathString: srcValue})
			//	}
			//}
		case atom.Link:
			// Добавляем обработку CSS файлов
			for _, a := range n.Attr {
				if a.Key == "href" && a.Val != "" &&
					(strings.HasSuffix(a.Val, ".css") || strings.Contains(strings.ToLower(getAttrValue(n, "rel")), "stylesheet")) {
					hrefValue := a.Val
					fmt.Println("stylesheet", hrefValue)
					resourceTags = append(
						resourceTags,
						ResourceTypes.StyleResource{SourceUrlString: url.String(), PathString: hrefValue})
				}
			}
		}
	}
	return resourceTags
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

func downloadResources(resourceTags []AnyResource, pathToSave url.URL) map[string]string {
	downloadedResources := map[string]string{}
	var wg sync.WaitGroup
	mu := sync.Mutex{}

	for _, resource := range resourceTags {
		wg.Add(1)
		go func(res AnyResource) {
			defer wg.Done()
			localURLString, err := ResourceTypes.DownloadResource(res, pathToSave)
			if err != nil {
				fmt.Println("Error downloading resource:", err)
				return
			}

			mu.Lock()
			// Используем полный Path() ресурса как ключ
			downloadedResources[res.Path()] = localURLString
			mu.Unlock()
			fmt.Println("Downloaded:", res.Path(), "->", localURLString)
		}(resource)
	}

	wg.Wait()
	return downloadedResources
}

func remapResourcePaths(downloadedResources map[string]string, htmlBody string) string {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		fmt.Println(err)
		return htmlBody
	}

	for n := range doc.Descendants() {
		if n.Type != html.ElementNode {
			continue
		}

		switch n.DataAtom {
		case atom.Img, atom.Script:
			for i, a := range n.Attr {
				if a.Key == "src" {
					resourcePath := a.Val

					// Проверяем, есть ли в нашей карте такой путь
					if localPath, ok := downloadedResources[resourcePath]; ok {
						fmt.Println("Mapped resource:", a.Val, "->", localPath)
						// Преобразуем путь для корректного отображения в HTML
						localPath = strings.ReplaceAll(localPath, "\\", "/")
						n.Attr[i].Val = "file:///" + localPath
					} else {
						fmt.Println("Warning: No local path found for:", resourcePath)
					}
				}
			}
		case atom.Link:
			// Добавляем обработку CSS ссылок
			for i, a := range n.Attr {
				if a.Key == "href" && (strings.HasSuffix(a.Val, ".css") ||
					strings.Contains(a.Val, "stylesheet")) {
					resourcePath := a.Val

					if localPath, ok := downloadedResources[resourcePath]; ok {
						fmt.Println("Mapped CSS resource:", a.Val, "->", localPath)
						localPath = strings.ReplaceAll(localPath, "\\", "/")
						n.Attr[i].Val = "file:///" + localPath
					}
				}
			}
		}
	}

	var buf bytes.Buffer
	err = html.Render(&buf, doc)
	if err != nil {
		fmt.Println("Error rendering HTML:", err)
		return htmlBody
	}

	return buf.String()
}

func main() {
	currentPath := "C:\\Users\\andre\\Desktop\\wb"
	currentPathURL, _ := url.Parse(currentPath)
	mainURL, _ := url.Parse("https://golang.org/")
	bodyResponse, _ := getHtml(mainURL.String())
	resources := extractResourceTags(mainURL, bodyResponse)
	downloadedResources := downloadResources(resources, *currentPathURL)

	remappedBody := remapResourcePaths(downloadedResources, bodyResponse)
	outputFile, _ := os.Create(filepath.Join(currentPath, "index.html"))
	outputFile.WriteString(remappedBody)

	fmt.Println(downloadedResources)
	outputFile.Close()
}
