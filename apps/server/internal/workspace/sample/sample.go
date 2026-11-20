package sample

import (
	_ "embed"
	"encoding/json"
	"time"
)

//go:embed workshop.json
var invitation []byte

//go:embed north-pier-registration.pdf
var PDF []byte

type Invitation struct {
	Subject, SenderAddress, BodyText, Title, TimeZone, Filename string
	SentAt, StartsAt, EndsAt                                    time.Time
	Preferences                                                 map[string]string
}

func Load() (v Invitation, err error) { err = json.Unmarshal(invitation, &v); return }
