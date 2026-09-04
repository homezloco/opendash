package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory      = 64 * 1024
	argonIterations  = 3
	argonParallelism = 2
	argonSaltLength  = 16
	argonKeyLength   = 32
)

var (
	ErrAdminExists        = errors.New("administrator already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("username or password does not meet requirements")
	ErrUnauthenticated    = errors.New("unauthenticated")
	dummyPasswordHash     string
)

func init() {
	// A fixed, syntactically valid hash makes unknown-user logins perform the same
	// expensive password operation as known-user logins without creating startup entropy dependencies.
	salt := make([]byte, argonSaltLength)
	hash := argon2.IDKey([]byte("invalid-password"), salt, argonIterations, argonMemory, argonParallelism, argonKeyLength)
	dummyPasswordHash = encodePasswordHash(salt, hash)
}

type Session struct {
	Username  string    `json:"username"`
	CSRFToken string    `json:"csrfToken"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Service struct {
	db       *sql.DB
	lifetime time.Duration
	now      func() time.Time
}

func New(db *sql.DB, lifetime time.Duration) *Service {
	return &Service{db: db, lifetime: lifetime, now: time.Now}
}

func (s *Service) BootstrapRequired(ctx context.Context) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM admins)").Scan(&exists)
	return !exists, err
}

func (s *Service) Bootstrap(ctx context.Context, username, password string) (string, Session, error) {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 64 || len(password) < 12 || len(password) > 1024 {
		return "", Session{}, ErrInvalidInput
	}
	hash, err := hashPassword(password)
	if err != nil {
		return "", Session{}, err
	}
	token, csrf, err := newSessionSecrets()
	if err != nil {
		return "", Session{}, err
	}
	now := s.now()
	expires := now.Add(s.lifetime)
	digest := sha256.Sum256([]byte(token))

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", Session{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO admins(id,username,password_hash,created_at) VALUES(1,?,?,?)", username, hash, now.Unix())
	if err != nil {
		return "", Session{}, err
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return "", Session{}, err
	}
	if inserted != 1 {
		return "", Session{}, ErrAdminExists
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO sessions(token_hash,admin_id,csrf_token,expires_at,created_at) VALUES(?,1,?,?,?)", digest[:], csrf, expires.Unix(), now.Unix()); err != nil {
		return "", Session{}, err
	}
	if err = tx.Commit(); err != nil {
		return "", Session{}, err
	}
	return token, Session{Username: username, CSRFToken: csrf, ExpiresAt: expires}, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (string, Session, error) {
	var stored, actual string
	err := s.db.QueryRowContext(ctx, "SELECT username,password_hash FROM admins WHERE username=?", strings.TrimSpace(username)).Scan(&actual, &stored)
	if errors.Is(err, sql.ErrNoRows) {
		stored = dummyPasswordHash
	} else if err != nil {
		return "", Session{}, err
	}
	valid := verifyPassword(password, stored)
	if errors.Is(err, sql.ErrNoRows) || !valid {
		return "", Session{}, ErrInvalidCredentials
	}

	token, csrf, err := newSessionSecrets()
	if err != nil {
		return "", Session{}, err
	}
	now := s.now()
	expires := now.Add(s.lifetime)
	digest := sha256.Sum256([]byte(token))
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", Session{}, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "DELETE FROM sessions WHERE admin_id=1 OR expires_at<=?", now.Unix()); err != nil {
		return "", Session{}, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO sessions(token_hash,admin_id,csrf_token,expires_at,created_at) VALUES(?,1,?,?,?)", digest[:], csrf, expires.Unix(), now.Unix()); err != nil {
		return "", Session{}, err
	}
	if err = tx.Commit(); err != nil {
		return "", Session{}, err
	}
	return token, Session{Username: actual, CSRFToken: csrf, ExpiresAt: expires}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (Session, error) {
	if token == "" {
		return Session{}, ErrUnauthenticated
	}
	digest := sha256.Sum256([]byte(token))
	var session Session
	var expires int64
	now := s.now().Unix()
	err := s.db.QueryRowContext(ctx, "SELECT a.username,s.csrf_token,s.expires_at FROM sessions s JOIN admins a ON a.id=s.admin_id WHERE s.token_hash=? AND s.expires_at>?", digest[:], now).Scan(&session.Username, &session.CSRFToken, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		_, _ = s.db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at<=?", now)
		return Session{}, ErrUnauthenticated
	}
	if err != nil {
		return Session{}, err
	}
	session.ExpiresAt = time.Unix(expires, 0).UTC()
	return session, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	digest := sha256.Sum256([]byte(token))
	_, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE token_hash=?", digest[:])
	return err
}

func newSessionSecrets() (string, string, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", "", err
	}
	csrf, err := randomToken(32)
	return token, csrf, err
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLength)
	return encodePasswordHash(salt, hash), nil
}

func encodePasswordHash(salt, hash []byte) string {
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argonMemory, argonIterations, argonParallelism, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash))
}

func verifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil ||
		memory != argonMemory || iterations != argonIterations || parallelism != argonParallelism {
		return false
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) != argonSaltLength {
		return false
	}
	expected, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(expected) != argonKeyLength {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}
