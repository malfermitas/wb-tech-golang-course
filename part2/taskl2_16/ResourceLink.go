package main

import (
	"net/url"

	"golang.org/x/net/html/atom"
)

type ResourceLink struct {
	linkType atom.Atom
	linkURL  url.URL
}

func NewResourceLink(linkType atom.Atom, linkURL url.URL) *ResourceLink {
	return &ResourceLink{linkType: linkType, linkURL: linkURL}
}

func (r ResourceLink) String() string {
	return r.linkURL.String()
}

func (r ResourceLink) LinkType() atom.Atom {
	return r.linkType
}

func (r ResourceLink) LinkURL() url.URL {
	return r.linkURL
}
