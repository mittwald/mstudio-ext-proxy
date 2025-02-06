package model

import "fmt"

type Session struct {
	ID            string `bson:"_id"`
	SessionSecret []byte
	UserID        string
	FirstName     string
	LastName      string
	Email         string
	AccessToken   string
}

func (s Session) CookieString() string {
	return fmt.Sprintf("%s:%X", s.ID, s.SessionSecret)
}
