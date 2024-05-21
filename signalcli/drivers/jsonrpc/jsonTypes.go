package signaljsonrpc

type jsonGroupInfo struct {
	GroupId string
	Type string
}

type jsonMsg struct {
	Destination string
	DestinationNumber string
	DestinationUuid string
	ExpiresInSeconds uint
	GroupInfo jsonGroupInfo
	Message string
	Timestamp uint64
	ViewOnce bool
}

type jsonSyncMsg struct {
	SentMessage jsonMsg
}

type jsonReceive struct {
	Account string
	Envelope jsonEnvelope
}

type jsonEnvelope struct {
	Source string
	SourceDevice uint
	SourceName string
	SourceNumber string
	SourceUuid string
	Timestamp uint64
	SyncMessage *jsonSyncMsg
}
