package auth

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
	"github.com/paveltessman/jao/users"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrAuthFailed   = errors.New("authentication failed")
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type UsersRepo interface {
	SelectByID(ctx context.Context, ID uuid.UUID) (users.User, error)
	Update(ctx context.Context, user users.User) error
}

type claims struct {
	jwt.RegisteredClaims
	UserID string `json:"userid"`
}

func newClaims(userID string, now time.Time, expiresAt time.Time, jti string) claims {
	c := claims{
		UserID:    userID,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    "jao",
		ID:        jti,
	}
	return c
}

type Tokens struct {
	Access  string
	Refresh string
}

type Service struct {
	users     UsersRepo
	secretKey []byte
}

func NewService(users UsersRepo, secretKey string) *Service {
	return &Service{users: users, secretKey: []byte(secretKey)}
}

func (s *Service) Authenticate(ctx context.Context, accessToken string) (users.User, error) {

	user, _, err := s.userFromToken(ctx, accessToken)
	if err != nil {
		return users.User{}, fmt.Errorf("%w: %w", ErrAuthFailed, err)
	}

	if !user.Active {
		return users.User{}, fmt.Errorf("%w: user is deactivated", ErrAuthFailed)
	}

	user.LastSeen = time.Now()

	err = s.users.Update(ctx, user)
	if err != nil {
		return users.User{}, fmt.Errorf("%w: unable to update last seen param for user: %w", ErrAuthFailed, err)
	}

	return user, nil
}

func (s *Service) Refresh(ctx context.Context, oldRefreshToken string) (Tokens, error) {

	user, oldClaims, err := s.userFromToken(ctx, oldRefreshToken)
	if err != nil {
		return Tokens{}, fmt.Errorf("%w: %w", ErrAuthFailed, err)
	}

	if oldClaims.ID != user.RefreshTokenID {
		return Tokens{}, fmt.Errorf("%w: refresh token id does not match provided jti", ErrAuthFailed)
	}

	if !user.Active {
		return Tokens{}, fmt.Errorf("%w: user is deactivated", ErrAuthFailed)
	}

	tokens, err := s.newTokenPair(ctx, user)

	return tokens, err

}

func (s *Service) userFromToken(ctx context.Context, tokenString string) (users.User, claims, error) {
	tokenClaims, err := s.parseToken(ctx, tokenString)
	if err != nil {
		return users.User{}, claims{}, err
	}

	userID, err := uuid.Parse(tokenClaims.UserID)
	if err != nil {
		return users.User{}, claims{}, fmt.Errorf("unable to parse user id from refresh token: %w", err)
	}

	user, err := s.users.SelectByID(ctx, userID)
	if err != nil {
		return users.User{}, claims{}, err
	}
	return user, tokenClaims, nil
}

func (s *Service) newTokenPair(ctx context.Context, user users.User) (Tokens, error) {

	accessToken, err := s.newToken(ctx, user.ID, accessTokenTTL, uuid.New())
	if err != nil {
		return Tokens{}, fmt.Errorf("unable to create access token: %w", err)
	}

	refreshTokenID := uuid.New()
	refreshToken, err := s.newToken(ctx, user.ID, refreshTokenTTL, refreshTokenID)
	if err != nil {
		return Tokens{}, fmt.Errorf("unable to create refresh token: %w", err)
	}

	tokens := Tokens{
		Access:  accessToken,
		Refresh: refreshToken,
	}

	user.RefreshTokenID = refreshTokenID.String()
	err = s.users.Update(ctx, user)
	if err != nil {
		return Tokens{}, fmt.Errorf("unable to update refresh token for user: %w", err)
	}

	return tokens, nil
}

func (s *Service) newToken(_ context.Context, userID uuid.UUID, timeDelta time.Duration, tokenID uuid.UUID) (string, error) {

	now := time.Now()
	expiresAt := now.Add(timeDelta)
	claims := newClaims(userID.String(), now, expiresAt, tokenID.String())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

func (s *Service) parseToken(_ context.Context, tokenString string) (claims, error) {

	var c claims
	token, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (any, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return claims{}, fmt.Errorf("%w: unable to parse token: %w", ErrInvalidToken, err)
	}

	if !token.Valid {
		return claims{}, ErrInvalidToken
	}

	return c, nil
}
