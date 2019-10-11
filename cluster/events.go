package cluster

type ModuleConnected struct {
	ModuleName string
}

type ApplyLogResponse struct {
	ApplyError string
	Comment    string
}

func (a ApplyLogResponse) Error() string {
	return a.ApplyError
}
