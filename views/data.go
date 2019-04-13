package views

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"

	// AlertMsgGeneric is displayed when a random error is encountered in out backend.
	AlertMsgGeneric = "Something went wrong. Please try again, and contact us if the problem persists."
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

// SetAlert takes an error an if it implements the PublicError interface, sets an alert msg to be the Public error,
// otherwise it sets the error msg to something generic.
func (d *Data) SetAlert(err error) {
	if pErr, ok := err.(PublicError); ok {
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: pErr.Public(),
		}
	} else {
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: AlertMsgGeneric,
		}
	}
}

type PublicError interface {
	error
	Public() string
}
