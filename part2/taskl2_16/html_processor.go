package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// updateHtmlLinks обновляет HTML, заменяя внешние ссылки на локальные пути
func updateHtmlLinks(resourceGraph *WebSiteResourceGraph, resourcesMap map[string]string, pathToSave url.URL) {
	var traverse func(*WebSiteResourceGraph)
	traverse = func(graph *WebSiteResourceGraph) {
		doc, err := html.Parse(strings.NewReader(graph.htmlBody))
		if err != nil {
			fmt.Printf("Ошибка парсинга HTML для %s: %v\n", graph.linkURL.String(), err)
			return
		}

		var processNode func(*html.Node)
		processNode = func(n *html.Node) {
			if n.Type == html.ElementNode {
				updateAttributes(n, resourcesMap, pathToSave)
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				processNode(c)
			}
		}
		processNode(doc)

		// Преобразуем обратно в строку
		var buf strings.Builder
		if err = html.Render(&buf, doc); err != nil {
			fmt.Printf("Ошибка рендеринга HTML для %s: %v\n", graph.linkURL.String(), err)
			return
		}
		graph.htmlBody = buf.String()

		// Рекурсивно обрабатываем потомков
		for i := range graph.Descendants() {
			traverse(&graph.Descendants()[i])
		}
	}

	traverse(resourceGraph)
}

// updateAttributes обновляет атрибуты узла
func updateAttributes(n *html.Node, resourcesMap map[string]string, pathToSave url.URL) {
	var attrKey string
	var needsHtmlExtension bool

	switch n.DataAtom {
	case atom.Img, atom.Script:
		attrKey = "src"
	case atom.Link:
		attrKey = "href"
	case atom.A:
		attrKey = "href"
		needsHtmlExtension = true
	default:
		return
	}

	for i, attr := range n.Attr {
		if attr.Key == attrKey {
			n.Attr[i].Val = convertToLocalPath(attr.Val, resourcesMap, pathToSave, needsHtmlExtension)
		}
	}
}

// convertToLocalPath преобразует URL в локальный путь
func convertToLocalPath(originalPath string, resourcesMap map[string]string, pathToSave url.URL, needsHtmlExtension bool) string {
	// Корневой путь -> index.html
	if originalPath == "/" {
		return "./index.html"
	}

	// Пути, начинающиеся с "/"
	if strings.HasPrefix(originalPath, "/") {
		newPath := "." + originalPath
		if needsHtmlExtension && !strings.Contains(filepath.Base(newPath), ".") {
			newPath += ".html"
		}
		return newPath
	}

	// Проверяем в resourcesMap
	if localPath, ok := resourcesMap[originalPath]; ok {
		relPath, err := filepath.Rel(pathToSave.String(), localPath)
		if err != nil {
			fmt.Printf("Ошибка получения относительного пути для %s: %v\n", localPath, err)
			return originalPath
		}
		return "./" + filepath.ToSlash(relPath)
	}

	// Проверяем абсолютные URL
	if parsedURL, err := url.Parse(originalPath); err == nil && parsedURL.IsAbs() {
		if parsedURL.Path == "/" {
			return "./index.html"
		}
		if localPath, ok := resourcesMap[parsedURL.String()]; ok {
			relPath, err := filepath.Rel(pathToSave.String(), localPath)
			if err != nil {
				fmt.Printf("Ошибка получения относительного пути для %s: %v\n", localPath, err)
				return originalPath
			}
			return "./" + filepath.ToSlash(relPath)
		}
	}

	// Проверяем относительные пути в resourcesMap
	for resourceURL, resourceLocalPath := range resourcesMap {
		if resourceParsedURL, parseErr := url.Parse(resourceURL); parseErr == nil {
			if resourceParsedURL.Path == originalPath {
				relPath, relErr := filepath.Rel(pathToSave.String(), resourceLocalPath)
				if relErr != nil {
					fmt.Printf("Ошибка получения относительного пути для %s: %v\n", resourceLocalPath, relErr)
					continue
				}
				return "./" + filepath.ToSlash(relPath)
			}
		}
	}

	return originalPath
}

// saveHtmlToFile сохраняет HTML в файл
func saveHtmlToFile(resourceGraph *WebSiteResourceGraph, pathToSave url.URL) {
	var traverse func(*WebSiteResourceGraph)
	traverse = func(graph *WebSiteResourceGraph) {
		// Формируем путь для сохранения HTML файла
		var filePath string

		// Проверяем, является ли это главной страницей (корневой путь)
		isMainPage := graph.linkURL.Path == "/" || graph.linkURL.Path == ""

		if isMainPage {
			// Для главной страницы сохраняем непосредственно в папке как index.html
			filePath = filepath.Join(pathToSave.String(), "index.html")
		} else {
			// Для остальных страниц формируем путь относительно pathToSave
			// Убираем ведущий слеш из пути, если он есть
			relativePath := strings.TrimPrefix(graph.linkURL.Path, "/")
			filePath = filepath.Join(pathToSave.String(), relativePath)

			// Если путь заканчивается на "/", добавляем index.html
			if strings.HasSuffix(filePath, string(os.PathSeparator)) {
				filePath = filepath.Join(filePath, "index.html")
			} else if !strings.HasSuffix(filePath, ".html") {
				// Если путь не заканчивается на .html, добавляем .html
				filePath += ".html"
			}
		}

		// Создаем директории при необходимости
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("Ошибка создания директории %s: %v\n", dir, err)
			return
		}

		// Открываем файл для записи
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Ошибка создания файла %s: %v\n", filePath, err)
			return
		}
		defer file.Close()

		// Записываем HTML тело в файл
		_, err = file.WriteString(graph.htmlBody)
		if err != nil {
			fmt.Printf("Ошибка записи в файл %s: %v\n", filePath, err)
			return
		}

		fmt.Printf("Сохранен файл: %s\n", filePath)

		// Рекурсивно обрабатываем потомков
		for i := range graph.Descendants() {
			traverse(&graph.Descendants()[i])
		}
	}

	traverse(resourceGraph)
}
