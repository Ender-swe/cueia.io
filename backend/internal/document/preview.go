package document

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type Signer struct {
	key []byte
	Now func() time.Time
}

func Load(ctx context.Context, db *pgxpool.Pool) (Signer, error) {
	var candidate [32]byte
	rand.Read(candidate[:])
	var key []byte
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "INSERT INTO deployment_signing_keys(singleton,key_material) VALUES(true,$1) ON CONFLICT DO NOTHING", candidate[:]); err != nil {
			return err
		}
		return tx.QueryRow(ctx, "SELECT key_material FROM deployment_signing_keys WHERE singleton").Scan(&key)
	})
	return Signer{key, time.Now}, err
}

type grant struct {
	Workspace, File, Purpose, Method string
	Expiry                           int64
}
type Link struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (s Signer) mac(payload string) []byte {
	h := hmac.New(sha256.New, s.key)
	h.Write([]byte(payload))
	return h.Sum(nil)
}
func (s Signer) Issue(w, id string) Link {
	expires := s.Now().UTC().Add(5 * time.Minute).Truncate(time.Second)
	claim, _ := json.Marshal(grant{w, id, "pdf-preview", "GET", expires.Unix()})
	payload := base64.RawURLEncoding.EncodeToString(claim)
	token := payload + "." + base64.RawURLEncoding.EncodeToString(s.mac(payload))
	return Link{"/api/previews/" + id + "?token=" + token, expires}
}

var ErrInvalid = errors.New("invalid preview")

func (s Signer) Redeem(id, method, token string) (string, error) {
	payload, sig, ok := strings.Cut(token, ".")
	if !ok {
		return "", ErrInvalid
	}
	signature, err := base64.RawURLEncoding.Strict().DecodeString(sig)
	if err != nil || !hmac.Equal(signature, s.mac(payload)) {
		return "", ErrInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", ErrInvalid
	}
	var g grant
	if json.Unmarshal(raw, &g) != nil || g.Workspace == "" || g.File != id || g.Purpose != "pdf-preview" || g.Method != method || s.Now().Unix() >= g.Expiry {
		return "", ErrInvalid
	}
	return g.Workspace, nil
}
