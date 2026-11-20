package document

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"testing"
	"time"
)

func TestGrantBounds(t *testing.T) {
	now := time.Now()
	s := Signer{[]byte("fixture key"), func() time.Time { return now }}
	link := s.Issue("owner", "document")
	u, _ := url.Parse(link.URL)
	token := u.Query().Get("token")
	for _, tc := range []struct{ id, method, token string }{
		{"other", "GET", token}, {"document", "POST", token}, {"document", "GET", token + "X"},
	} {
		if _, err := s.Redeem(tc.id, tc.method, tc.token); err == nil {
			t.Fatal("accepted invalid grant")
		}
	}
	raw, _ := json.Marshal(grant{"owner", "document", "download", "GET", link.ExpiresAt.Unix()})
	payload := base64.RawURLEncoding.EncodeToString(raw)
	signed := payload + "." + base64.RawURLEncoding.EncodeToString(s.mac(payload))
	if _, err := s.Redeem("document", "GET", signed); err == nil {
		t.Fatal("accepted wrong purpose")
	}
	now = link.ExpiresAt
	if _, err := s.Redeem("document", "GET", token); err == nil {
		t.Fatal("accepted expired grant")
	}
}
