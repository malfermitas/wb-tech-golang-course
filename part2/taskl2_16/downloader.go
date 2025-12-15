package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func downloadResources(resourceGraph *WebSiteResourceGraph, pathToSave url.URL) map[string]string {
	var resources = make(map[string]string)

	var traverse func(*WebSiteResourceGraph)
	traverse = func(graph *WebSiteResourceGraph) {
		for _, resourceLink := range graph.ResourceLinks() {
			resourceURL := resourceLink.LinkURL()
			filePath := filepath.Join(pathToSave.String(), resourceURL.Path)
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				fmt.Printf("Ошибка создания директории %s: %v\n", dir, err)
				continue
			}

			file, err := os.Create(filePath)
			if err != nil {
				fmt.Printf("Ошибка создания файла %s: %v\n", filePath, err)
				continue
			}

			client := &http.Client{
				Timeout: 10 * time.Second,
			}
			resp, err := client.Get(resourceURL.String())
			if err != nil {
				fmt.Printf("Ошибка загрузки ресурса %s: %v\n", resourceURL.String(), err)
				continue
			}

			_, err = io.Copy(file, resp.Body)
			if err != nil {
				fmt.Printf("Ошибка записи в файл %s: %v\n", filePath, err)
				continue
			}

			resources[resourceURL.String()] = filePath
			file.Close()
			resp.Body.Close()
		}

		for i := range graph.Descendants() {
			traverse(&graph.Descendants()[i])
		}
	}

	traverse(resourceGraph)

	return resources
}

// ProcessWebsite полностью обрабатывает веб-сайт: краулинг, скачивание ресурсов, обновление ссылок и сохранение
func ProcessWebsite(mainURL url.URL, pathToSave url.URL) {
	fmt.Println("Начинаем обработку веб-сайта...")

	// 1. Краулинг сайта
	fmt.Println("Краулинг сайта...")
	resourceGraph := CrawlWebSite(mainURL)

	// 2. Скачивание ресурсов
	fmt.Println("Скачивание ресурсов...")
	resourcesMap := downloadResources(resourceGraph, pathToSave)

	// 3. Обновление HTML ссылок
	fmt.Println("Обновление HTML ссылок...")
	updateHtmlLinks(resourceGraph, resourcesMap, pathToSave)

	// 4. Сохранение HTML файлов
	fmt.Println("Сохранение HTML файлов...")
	saveHtmlToFile(resourceGraph, pathToSave)

	fmt.Println("Обработка веб-сайта завершена!")
}
