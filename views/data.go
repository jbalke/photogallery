package views

import (
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"lenslocked.com/models"
)

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"

	// AlertMsgGeneric is displayed when a random error is encountered in out backend.
	AlertMsgGeneric = "Something went wrong. Please try again, and contact us if the problem persists."
	// AlertLevel is the name of our cookie for persisting alert level for flash alerts.
	AlertLevel = "alert_level"
	// AlertMessage is the name of our cookie for persisting alert messages for flash alerts.
	AlertMessage = "alert_message"
)

// Alert is used to render Bootstrap alert messages in templates
type Alert struct {
	Level   string
	Message string
}

// Data is the data structure that views expect template data to be passed in
type Data struct {
	Alert *Alert
	User  *models.User
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
		log.Println(err)
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: AlertMsgGeneric,
		}
	}
}

func AlertSuccess(msg string) Alert {
	return Alert{
		Level:   AlertLvlSuccess,
		Message: msg,
	}
}

func (d *Data) AlertError(msg string) {
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}

type PublicError interface {
	error
	Public() string
}

func persistAlert(w http.ResponseWriter, alert Alert) {
	expiresAt := time.Now().Add(time.Minute)
	lvl := http.Cookie{
		Name:     AlertLevel,
		Value:    encode(alert.Level),
		Expires:  expiresAt,
		HttpOnly: true,
	}
	msg := http.Cookie{
		Name:     AlertMessage,
		Value:    encode(alert.Message),
		Expires:  expiresAt,
		HttpOnly: true,
	}
	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}

func clearAlert(w http.ResponseWriter) {
	expiresAt := time.Now()
	lvl := http.Cookie{
		Name:     AlertLevel,
		Value:    "",
		Expires:  expiresAt,
		HttpOnly: true,
	}
	msg := http.Cookie{
		Name:     AlertMessage,
		Value:    "",
		Expires:  expiresAt,
		HttpOnly: true,
	}
	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}

func getAlert(r *http.Request) *Alert {
	lvl, err := r.Cookie(AlertLevel)
	if err != nil {
		return nil
	}
	msg, err := r.Cookie(AlertMessage)
	if err != nil {
		return nil
	}
	return &Alert{
		Level:   decode(lvl.Value),
		Message: decode(msg.Value),
	}
}

func RedirectAlert(w http.ResponseWriter, r *http.Request, urlString string, code int, alert Alert) {
	persistAlert(w, alert)
	http.Redirect(w, r, urlString, code)
}

func encode(src string) string {
	return base64.URLEncoding.EncodeToString([]byte(src))
}

func decode(src string) string {
	bytes, _ := base64.URLEncoding.DecodeString(src)
	return string(bytes)
}
