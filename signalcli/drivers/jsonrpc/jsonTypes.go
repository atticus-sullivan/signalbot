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
	Previews []jsonPreview
	TextStyles []jsonTextStyle
	RemoteDelete struct {
		Timestamp uint64
	}
	Quote jsonQuote
	Reaction jsonReaction
}

type jsonReaction struct {
	Emoji string
	TargetAuthor string
	TargetAuthorNumber string
	TargetAuthorUuid string
	TargetSentTimestamp string
	IsRemove bool
}

type jsonTextStyle struct {
	Style string
	Start uint
	Length uint
}

type jsonPreview struct {
	Url string
	Title string
	Description string
	Image jsonImage
}

type jsonImage struct {
	ContentType string
	Filename string
	Id string
	Size uint
	Width uint
	Height uint
	Caption string
	UploadTimestamp uint64
}

type jsonSyncMsg struct {
	SentMessage jsonMsg
	ReadMessages []jsonRead
}

type jsonRead struct {
	Sender string
	SenderNumber string
	SenderUuid string
	Timestamp uint64
}
type jsonTypingMsg struct {
	Action string
	Timestamp uint64
	GroupId string
}

type jsonReceive struct {
	Account string
	Envelope jsonEnvelope
	Exception *jsonException
}

type jsonException struct {
	Message string
	Type string
}

type jsonQuote struct {
	Attachments []string
	Author string
	AuthorNumber string
	AuthorUuid string
	Id uint64
	Text string
}

type jsonAttachment struct {
	ContentType string
	Filename string
	Id string
	Size uint
	Width uint
	Height uint
	Caption string
	UploadTimestamp uint64
}

type jsonDataMsg struct {
	Attachments []jsonAttachment
	ExpiresInSeconds uint
	GroupInfo jsonGroupInfo
	Message string
	Quote jsonQuote
	Timestamp uint64
	ViewOnce bool
}

type jsonReceiptMsg struct {
	When uint64
	IsDelivery bool
	IsRead bool
	IsViewed bool
	Timestamps []uint64
}

type jsonEnvelope struct {
	Source string
	SourceDevice uint
	SourceName string
	SourceNumber string
	SourceUuid string
	Timestamp uint64
	SyncMessage *jsonSyncMsg
	DataMessage *jsonDataMsg
	TypingMessage *jsonTypingMsg
	ReceiptMessage *jsonReceiptMsg
}
