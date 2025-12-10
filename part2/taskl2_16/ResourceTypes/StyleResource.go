package ResourceTypes

type StyleResource struct {
	SourceUrlString string
	PathString      string
}

func (s StyleResource) RemoteUrlString() string {
	return s.SourceUrlString
}

func (s StyleResource) Path() string {
	return s.PathString
}
