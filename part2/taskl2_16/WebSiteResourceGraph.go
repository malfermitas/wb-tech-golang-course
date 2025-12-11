package main

import (
	"net/url"
)

// WebSiteResourceGraph this structure represents a single web page and the links to resources it contains
type WebSiteResourceGraph struct {
	linkURL       url.URL
	htmlBody      string
	resourceLinks []ResourceLink
	descendants   []WebSiteResourceGraph
}

func NewWebSiteResourceGraph(
	linkURL url.URL,
	resourceLinks []ResourceLink,
	htmlBody string,
) *WebSiteResourceGraph {
	return &WebSiteResourceGraph{linkURL: linkURL, resourceLinks: resourceLinks, htmlBody: htmlBody}
}

func (graph *WebSiteResourceGraph) LinkURL() url.URL {
	return graph.linkURL
}

func (graph *WebSiteResourceGraph) ResourceLinks() []ResourceLink {
	return graph.resourceLinks
}

func (graph *WebSiteResourceGraph) Descendants() []WebSiteResourceGraph {
	return graph.descendants
}

func (graph *WebSiteResourceGraph) AddDescendant(descendant WebSiteResourceGraph) {
	graph.descendants = append(graph.descendants, descendant)
}
