package cluster

const (
	ApplyCommandEvent = "CONFIG_CLUSTER:APPLY_COMMAND"

	TokenParam   = "token"
	ClusterParam = "cluster"
)
const (
	_ = iota
	BackendDeclarationCommand
	ConfigSchemaCommand
)
