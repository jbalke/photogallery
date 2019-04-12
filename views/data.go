package views

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"
)

// Alert is used to render Bootstrap alert messages in templates
type Alert struct {
	Level   string
	Message string
}

// Data is the data structure that views expect template data to be passed in
type Data struct {
	Alert *Alert
	Yield interface{}
}
