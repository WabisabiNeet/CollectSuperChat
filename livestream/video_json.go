package livestream

import ()

// LiveChatInfo is json struct
type LiveChatInfo struct {
	Contents *Contents `json:"contents,omitempty"`
}

// Contents is json struct
type Contents struct {
	TwoColumnWatchNextResults *TwoColumnWatchNextResults `json:"twoColumnWatchNextResults,omitempty"`
}

// TwoColumnWatchNextResults is json struct
type TwoColumnWatchNextResults struct {
	ConversationBar *ConversationBar `json:"conversationBar,omitempty"`
}

// ConversationBar is json struct
type ConversationBar struct {
	LiveChatRenderer *LiveChatRenderer `json:"liveChatRenderer,omitempty"`
}

// LiveChatRenderer is json struct
type LiveChatRenderer struct {
	Continuations []*Continuation `json:"continuations,omitempty"`
}

// Continuation is json struct
type Continuation struct {
	ReloadContinuationData *ReloadContinuationData `json:"reloadContinuationData,omitempty"`
}

// ReloadContinuationData is json struct
type ReloadContinuationData struct {
	Continuation string `json:"continuation,omitempty"`
}
