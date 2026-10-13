package credentials

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("credential not found")

type Credential struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Header    string    `json:"header"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Input struct {
	Name   string `json:"name"`
	Host   string `json:"host"`
	Header string `json:"header"`
	Value  string `json:"value"`
}

func (i *Input) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.Host = strings.ToLower(strings.TrimSpace(i.Host))
	i.Header = http.CanonicalHeaderKey(strings.TrimSpace(i.Header))
	u, err := url.Parse("https://" + i.Host)
	if i.Name == "" || len(i.Name) > 100 || err != nil || u.Host != i.Host || u.Hostname() == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("name and an exact destination host are required")
	}
	if i.Header == "" || len(i.Header) > 100 {
		return fmt.Errorf("a valid header name is required")
	}
	for _, c := range i.Header {
		if !strings.ContainsRune("!#$%&'*+-.^_`|~0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", c) {
			return fmt.Errorf("invalid header name")
		}
	}
	switch i.Header {
	case "Host", "Content-Length", "Connection", "Transfer-Encoding", "Idempotency-Key", "Cookie", "Proxy-Authorization", "Proxy-Connection", "Trailer", "Te", "Upgrade":
		return fmt.Errorf("this header cannot store a credential")
	}
	if i.Value == "" || len(i.Value) > 8192 {
		return fmt.Errorf("credential value must contain 1–8192 bytes")
	}
	for _, c := range []byte(i.Value) {
		if c < 32 || c > 126 {
			return fmt.Errorf("credential value must contain printable ASCII characters")
		}
	}
	return nil
}

type Vault struct {
	pool *pgxpool.Pool
	aead cipher.AEAD
}

func New(pool *pgxpool.Pool, key string) (*Vault, error) {
	decoded, err := hex.DecodeString(key)
	if err != nil || len(decoded) != 32 {
		return nil, fmt.Errorf("CREDENTIALS_ENCRYPTION_KEY must be 64 hexadecimal characters")
	}
	block, err := aes.NewCipher(decoded)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Vault{pool: pool, aead: aead}, nil
}

func (v *Vault) seal(owner, id uuid.UUID, value string) []byte {
	nonce := make([]byte, v.aead.NonceSize())
	_, _ = rand.Read(nonce)
	return v.aead.Seal(nonce, nonce, []byte(value), []byte(owner.String()+id.String()))
}

func (v *Vault) open(owner, id uuid.UUID, data []byte) (string, error) {
	if len(data) < v.aead.NonceSize() {
		return "", errors.New("invalid encrypted credential")
	}
	plain, err := v.aead.Open(nil, data[:v.aead.NonceSize()], data[v.aead.NonceSize():], []byte(owner.String()+id.String()))
	if err != nil {
		return "", errors.New("unable to decrypt credential")
	}
	return string(plain), nil
}

func (v *Vault) List(ctx context.Context, owner uuid.UUID) ([]Credential, error) {
	rows, err := v.pool.Query(ctx, `SELECT id,name,host,header,updated_at FROM credentials WHERE owner_id=$1 ORDER BY name,id`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Credential{}
	for rows.Next() {
		var c Credential
		if err := rows.Scan(&c.ID, &c.Name, &c.Host, &c.Header, &c.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// Save replaces a value without ever returning its plaintext or ciphertext.
func (v *Vault) Save(ctx context.Context, owner, id uuid.UUID, input Input) (*Credential, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	c := &Credential{ID: id, Name: input.Name, Host: input.Host, Header: input.Header}
	if id == uuid.Nil {
		c.ID = uuid.New()
		err := v.pool.QueryRow(ctx, `INSERT INTO credentials(id,owner_id,name,host,header,encrypted_value) VALUES($1,$2,$3,$4,$5,$6) RETURNING updated_at`, c.ID, owner, c.Name, c.Host, c.Header, v.seal(owner, c.ID, input.Value)).Scan(&c.UpdatedAt)
		return c, err
	}
	err := v.pool.QueryRow(ctx, `UPDATE credentials SET name=$3,host=$4,header=$5,encrypted_value=$6,updated_at=clock_timestamp() WHERE id=$1 AND owner_id=$2 RETURNING updated_at`, id, owner, c.Name, c.Host, c.Header, v.seal(owner, id, input.Value)).Scan(&c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (v *Vault) Delete(ctx context.Context, owner, id uuid.UUID) error {
	tag, err := v.pool.Exec(ctx, `DELETE FROM credentials WHERE id=$1 AND owner_id=$2`, id, owner)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// Resolve scopes credentials to the owner admitted with the run and one HTTPS host.
func (v *Vault) Resolve(ctx context.Context, owner, id uuid.UUID, destination *url.URL) (string, string, error) {
	var host, header string
	var data []byte
	err := v.pool.QueryRow(ctx, `SELECT host,header,encrypted_value FROM credentials WHERE id=$1 AND owner_id=$2`, id, owner).Scan(&host, &header, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrNotFound
	}
	if err != nil {
		return "", "", errors.New("credential unavailable")
	}
	if destination.Scheme != "https" || !strings.EqualFold(destination.Host, host) {
		return "", "", errors.New("credential requires its configured HTTPS host")
	}
	value, err := v.open(owner, id, data)
	return header, value, err
}

type Resolver func(context.Context, uuid.UUID, *url.URL) (string, string, error)
type resolverKey struct{}
type ownerKey struct{}

func WithOwner(ctx context.Context, owner uuid.UUID) context.Context {
	return context.WithValue(ctx, ownerKey{}, owner)
}
func Owner(ctx context.Context) uuid.UUID {
	owner, _ := ctx.Value(ownerKey{}).(uuid.UUID)
	return owner
}
func WithResolver(ctx context.Context, resolver Resolver) context.Context {
	return context.WithValue(ctx, resolverKey{}, resolver)
}
func FromContext(ctx context.Context) Resolver {
	resolver, _ := ctx.Value(resolverKey{}).(Resolver)
	return resolver
}
