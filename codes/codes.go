package codes

// TODO numbers
const (
	// Common
	StartHttpServerError    = 0
	ShutdownHttpServerError = 0
	ShutdownHttpServerInfo  = 0

	ConfigDataSerializeError = 0
	SocketIoError            = 0

	// Raft
	InitRaftError              = 0
	RestoreFromRepositoryError = 0
	RaftLoggerCode             = 0
	BootstrapClusterError      = 0
	RaftShutdownError          = 0
	RaftShutdownInfo           = 0

	// Raft Log
	InvalidLogDataCommand  = 0
	ApplyLogCommandError   = 0
	UnknownLogCommand      = 0
	PrepareLogCommandError = 0
	SyncApplyError         = 0

	// Raft Cluster
	ConnectToLeaderError          = 0
	SendDeclarationToLeaderError  = 0
	LeaderClientReconnectionError = 0
	LeaderClientDisconnected      = 0
	LeaderClientBindError         = 0

	// Services
	DiscoveryServiceError = 0
	RoutesServiceError    = 0

	DeleteBackendDeclarationError = 0
)
