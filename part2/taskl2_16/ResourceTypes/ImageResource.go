package ResourceTypes

type ImageResource struct {
	SourceUrlString string
	PathString      string
}

func (i ImageResource) RemoteUrlString() string {
	return i.SourceUrlString
}

func (i ImageResource) Path() string {
	return i.PathString
}
