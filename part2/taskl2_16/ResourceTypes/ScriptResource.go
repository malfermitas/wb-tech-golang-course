package ResourceTypes

type ScriptResource struct {
	SourceUrlString string
	PathString      string
}

func (s ScriptResource) RemoteUrlString() string {
	return s.SourceUrlString
}

func (s ScriptResource) Path() string {
	return s.PathString
}
