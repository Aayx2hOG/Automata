package services

import (
	"context"
	"errors"

	"github.com/Aayx2hOG/automata/internal/models"
	clerk "github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	clerkjwt "github.com/clerk/clerk-sdk-go/v2/jwt"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ClerkAuth maps verified Clerk subjects to local users without linking accounts by email.
type ClerkAuth struct {
	pool    *pgxpool.Pool
	keys    *jwks.Client
	users   *clerkuser.Client
	origins map[string]bool
}

func NewClerkAuth(pool *pgxpool.Pool, secret string, origins []string) *ClerkAuth {
	cfg := &clerk.ClientConfig{BackendConfig: clerk.BackendConfig{Key: clerk.String(secret)}}
	allowed := make(map[string]bool, len(origins))
	for _, origin := range origins {
		allowed[origin] = true
	}
	return &ClerkAuth{pool: pool, keys: jwks.NewClient(cfg), users: clerkuser.NewClient(cfg), origins: allowed}
}

func (s *ClerkAuth) Authenticate(ctx context.Context, token string) (*models.User, error) {
	claims, err := clerkjwt.Verify(ctx, &clerkjwt.VerifyParams{
		Token: token, JWKSClient: s.keys,
		AuthorizedPartyHandler: func(origin string) bool { return s.origins[origin] },
	})
	if err != nil || claims.Subject == "" || claims.SessionID == "" {
		return nil, models.ErrorInvalidToken
	}
	u, err := s.lookup(ctx, claims.Subject)
	if err == nil {
		if !u.IsActive {
			return nil, models.ErrorInvalidCredentials
		}
		return u, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	profile, err := s.users.Get(ctx, claims.Subject)
	if err != nil {
		return nil, err
	}
	email, err := verifiedPrimaryEmail(profile)
	if err != nil {
		return nil, err
	}
	// Never trust client metadata for roles or attach an existing account by email.
	// A duplicate email requires an explicit account migration by the administrator.
	_, err = s.pool.Exec(ctx, `INSERT INTO users (username,email,password_hash,role,is_active,clerk_user_id)
 VALUES ($1,$2,'!clerk','user',true,$3) ON CONFLICT (clerk_user_id) DO NOTHING`,
		"user_"+uuid.NewString(), email, claims.Subject)
	if err != nil {
		return nil, err
	}
	u, err = s.lookup(ctx, claims.Subject)
	if err == nil && !u.IsActive {
		return nil, models.ErrorInvalidCredentials
	}
	return u, err
}

func verifiedPrimaryEmail(profile *clerk.User) (string, error) {
	if profile == nil || profile.Banned || profile.Locked || profile.PrimaryEmailAddressID == nil {
		return "", models.ErrorInvalidCredentials
	}
	for _, email := range profile.EmailAddresses {
		if email != nil && email.ID == *profile.PrimaryEmailAddressID && email.Verification != nil && email.Verification.Status == "verified" {
			return email.EmailAddress, nil
		}
	}
	return "", models.ErrorInvalidCredentials
}

func (s *ClerkAuth) lookup(ctx context.Context, subject string) (*models.User, error) {
	u := &models.User{}
	err := s.pool.QueryRow(ctx, `SELECT id,username,email,role,is_active,created_at,updated_at FROM users WHERE clerk_user_id=$1`, subject).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}
