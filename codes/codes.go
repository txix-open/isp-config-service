package codes

// TODO numbers
const (
	// Common
	StartHttpServerError    = 0
	ShutdownHttpServerError = 0
	ShutdownHttpServerInfo  = 0

	SocketIoError = 0

	// Raft
	InitRaftError              = 0
	RestoreFromRepositoryError = 0
	RaftLoggerCode             = 0
	BootstrapClusterError      = 0
	RaftShutdownError          = 0
	RaftShutdownInfo           = 0

	// Raft Log
	ApplyLogCommandError   = 0
	PrepareLogCommandError = 0
	SyncApplyError         = 0

	// Raft Cluster
	SendDeclarationToLeaderError = 0
	LeaderClientConnectionError  = 0
	LeaderClientDisconnected     = 0

	// Services
	DiscoveryServiceSendModulesError = 0
	RoutesServiceSendRoutesError     = 0

	DeleteBackendDeclarationError = 0
)
