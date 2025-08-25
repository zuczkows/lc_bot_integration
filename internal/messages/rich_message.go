package messages

type RichMessage struct {
	Type       string    `json:"type"`
	TemplateID string    `json:"template_id"`
	Elements   []Element `json:"elements"`
}

type Element struct {
	Title   string   `json:"title"`
	Buttons []Button `json:"buttons"`
}

type Button struct {
	Type       string   `json:"type"`
	Text       string   `json:"text"`
	PostbackID string   `json:"postback_id"`
	UserIDs    []string `json:"user_ids"`
	Value      string   `json:"value"`
}

func NewQuickReplies(title string, buttons []Button) RichMessage {
	return RichMessage{
		Type:       "rich_message",
		TemplateID: "quick_replies",
		Elements: []Element{
			{
				Title:   title,
				Buttons: buttons,
			},
		},
	}
}
