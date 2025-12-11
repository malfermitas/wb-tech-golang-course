package main

import (
	"net/url"
	"os"
)

func main() {
	mainURLStr := os.Args[1]
	mainURL, _ := url.Parse(mainURLStr)
	desktopURL, _ := url.Parse(os.Args[2])

	ProcessWebsite(*mainURL, *desktopURL)
}
