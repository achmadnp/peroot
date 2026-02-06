package model

// WhatsApp Cloud API request/response models.
// Reference: https://developers.facebook.com/docs/whatsapp/cloud-api/reference/messages

type WAMessageRequest struct {
	MessagingProduct string      `json:"messaging_product"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Template         *WATemplate `json:"template,omitempty"`
}

type WATemplate struct {
	Name       string        `json:"name"`
	Language   *WALanguage   `json:"language"`
	Components []WAComponent `json:"components,omitempty"`
}

type WALanguage struct {
	Code string `json:"code"`
}

type WAComponent struct {
	Type       string        `json:"type"`
	Parameters []WAParameter `json:"parameters"`
}

type WAParameter struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type WAMessageResponse struct {
	MessagingProduct string      `json:"messaging_product"`
	Contacts         []WAContact `json:"contacts"`
	Messages         []WAMsg     `json:"messages"`
}

type WAContact struct {
	Input string `json:"input"`
	WaID  string `json:"wa_id"`
}

type WAMsg struct {
	ID string `json:"id"`
}

type WAErrorResponse struct {
	Error WAErrorDetail `json:"error"`
}

type WAErrorDetail struct {
	Message   string `json:"message"`
	Type      string `json:"type"`
	Code      int    `json:"code"`
	FBTraceID string `json:"fbtrace_id"`
}
