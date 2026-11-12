package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
	"github.com/paveltessman/jao/users"
)

var _ UsersRepo = (*fakeUsersRepo)(nil)

type fakeUsersRepo struct {
	onSelectUser    users.User
	errOnSelectUser error
	onUpdateErr     error
	updateCallArgs  []users.User
}

func (r *fakeUsersRepo) SelectByID(ctx context.Context, ID uuid.UUID) (users.User, error) {
	return r.onSelectUser, r.errOnSelectUser
}

func (r *fakeUsersRepo) Update(ctx context.Context, user users.User) error {
	r.updateCallArgs = append(r.updateCallArgs, user)
	return r.onUpdateErr
}

func userIDFromTokenString(t *testing.T, tokenString string, svc *Service) string {
	t.Helper()

	c, err := svc.parseToken(t.Context(), tokenString)
	if err != nil {
		t.Fatalf("error when parsing token: %v", err)
	}

	return c.UserID

}

func expiredTokenForUser(t *testing.T, svc *Service, userID uuid.UUID, tokenID uuid.UUID) string {
	t.Helper()
	return tokenForUser(t, svc, userID, tokenID, 0)
}

func validTokenForUser(t *testing.T, svc *Service, userID uuid.UUID, tokenID uuid.UUID) string {
	t.Helper()
	return tokenForUser(t, svc, userID, tokenID, 100*time.Hour)
}

func tokenForUser(t *testing.T, svc *Service, userID uuid.UUID, tokenID uuid.UUID, ttl time.Duration) string {
	t.Helper()

	tokenString, err := svc.newToken(t.Context(), userID, ttl, tokenID)
	if err != nil {
		t.Fatalf("unable to create new token: %v", err)
	}
	return tokenString
}

func readClaims(t *testing.T, svc *Service, tokenString string) claims {
	t.Helper()

	c, err := svc.parseToken(t.Context(), tokenString)
	if err != nil {
		t.Fatalf("error when parsing token: %v", err)
	}
	return c
}

func tokenContainsUserID(t *testing.T, svc *Service, tokenString string, expected uuid.UUID) bool {
	t.Helper()
	got := userIDFromTokenString(t, tokenString, svc)
	return got == expected.String()
}

func readTokenID(t *testing.T, svc *Service, tokenString string) string {
	c := readClaims(t, svc, tokenString)
	return c.ID
}

func readTokenDuration(t *testing.T, svc *Service, tokenString string) time.Duration {
	t.Helper()

	c := readClaims(t, svc, tokenString)
	return c.ExpiresAt.Sub(c.IssuedAt.Time)
}

func newUser() users.User {
	u := users.User{
		ID: uuid.New(),
	}
	return u
}

func TestAuthenticateWithValidToken(t *testing.T) {

	expectedUser := newUser()
	expectedUser.Active = true
	usersRepo := &fakeUsersRepo{
		onSelectUser: expectedUser,
	}
	svc := NewService(usersRepo, "opa")

	loadedUser, err := svc.Authenticate(t.Context(), validTokenForUser(t, svc, expectedUser.ID, uuid.New()))
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}

	expectedUser.LastSeen = loadedUser.LastSeen

	if loadedUser != expectedUser {
		t.Error("users do not match")
	}
}

func TestAuthenticateWithExpiredToken(t *testing.T) {

	expectedUser := newUser()
	expectedUser.Active = true
	usersRepo := &fakeUsersRepo{
		onSelectUser: expectedUser,
	}
	svc := NewService(usersRepo, "opa")

	_, err := svc.Authenticate(t.Context(), expiredTokenForUser(t, svc, expectedUser.ID, uuid.New()))
	if !errors.Is(err, ErrAuthFailed) && !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected auth error, got something else: %v", err)
	}
}

func TestAuthenticateWithInvalidToken(t *testing.T) {

	expectedUser := newUser()
	expectedUser.Active = true
	usersRepo := &fakeUsersRepo{
		onSelectUser: expectedUser,
	}
	svc := NewService(usersRepo, "opa")

	_, err := svc.Authenticate(t.Context(), "opa opa")
	if !errors.Is(err, ErrAuthFailed) && !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("expected auth error, got something else: %v", err)
	}
}

func TestAuthenticateWithDeactivatedUser(t *testing.T) {

	expectedUser := newUser()
	expectedUser.Active = false
	usersRepo := &fakeUsersRepo{
		onSelectUser: expectedUser,
	}
	svc := NewService(usersRepo, "opa")

	_, err := svc.Authenticate(t.Context(), validTokenForUser(t, svc, expectedUser.ID, uuid.New()))
	if !errors.Is(err, ErrAuthFailed) && !strings.Contains(err.Error(), "deactivated") {
		t.Fatalf("expected auth error, got something else: %v", err)
	}
}

func TestAuthenticateWithWithNonexistentUser(t *testing.T) {

	dbError := errors.New("opa x2")

	usersRepo := &fakeUsersRepo{
		onSelectUser:    users.User{},
		errOnSelectUser: dbError,
	}
	svc := NewService(usersRepo, "opa")

	_, err := svc.Authenticate(t.Context(), validTokenForUser(t, svc, uuid.New(), uuid.New()))
	if !errors.Is(err, ErrAuthFailed) && !strings.Contains(err.Error(), "opa x2") {
		t.Fatalf("expected auth error, got something else: %v", err)
	}
}

func TestAuthenticateUpdatesLastSeen(t *testing.T) {

	expectedUser := newUser()
	expectedUser.Active = true
	expectedUser.LastSeen = time.Now()
	usersRepo := &fakeUsersRepo{
		onSelectUser: expectedUser,
	}
	svc := NewService(usersRepo, "opa")

	loadedUser, err := svc.Authenticate(t.Context(), validTokenForUser(t, svc, expectedUser.ID, uuid.New()))
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}

	if !loadedUser.LastSeen.After(expectedUser.LastSeen) {
		t.Error("last seen parameter is incorrect")
	}

	if len(usersRepo.updateCallArgs) != 1 {
		t.Fatalf("user was not updated")
	}

	if callArg := usersRepo.updateCallArgs[0]; !callArg.LastSeen.Equal(loadedUser.LastSeen) {
		t.Fatalf("user update was called with incorrect last seen argument")
	}
}

func TestAuthenticateHandlesErrorOnLastSeenUpdate(t *testing.T) {

	dbError := errors.New("fuk")

	expectedUser := newUser()
	expectedUser.Active = true
	expectedUser.LastSeen = time.Now()
	usersRepo := &fakeUsersRepo{
		onSelectUser: expectedUser,
		onUpdateErr:  dbError,
	}
	svc := NewService(usersRepo, "opa")

	_, err := svc.Authenticate(t.Context(), validTokenForUser(t, svc, expectedUser.ID, uuid.New()))
	if !errors.Is(err, ErrAuthFailed) && !strings.Contains(err.Error(), "fuk") {
		t.Fatalf("expected auth error, got something else: %v", err)
	}
}

func TestRefreshWithValidToken(t *testing.T) {

	refreshTokenID := uuid.New()
	user := newUser()
	user.Active = true
	user.RefreshTokenID = refreshTokenID.String()
	usersRepo := &fakeUsersRepo{
		onSelectUser: user,
	}
	svc := NewService(usersRepo, "opa")

	_, err := svc.Refresh(t.Context(), validTokenForUser(t, svc, user.ID, refreshTokenID))
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}
}

func TestRefreshWithExpiredToken(t *testing.T) {

	user := newUser()
	usersRepo := &fakeUsersRepo{
		onSelectUser: user,
	}
	svc := NewService(usersRepo, "opa")

	_, err := svc.Refresh(t.Context(), expiredTokenForUser(t, svc, user.ID, uuid.New()))
	if !errors.Is(err, ErrAuthFailed) && !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected auth error, got something else: %v", err)
	}
}

func TestRefreshWithInvalidToken(t *testing.T) {

	refreshTokenID := uuid.New()
	user := newUser()
	user.Active = false
	user.RefreshTokenID = refreshTokenID.String()
	usersRepo := &fakeUsersRepo{
		onSelectUser: user,
	}
	svc := NewService(usersRepo, "opa")

	_, err := svc.Refresh(t.Context(), "opa opa")
	if !errors.Is(err, ErrAuthFailed) && !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("expected auth error, got something else: %v", err)
	}
}

func TestRefreshWithDeactivatedUser(t *testing.T) {

	refreshTokenID := uuid.New()
	user := newUser()
	user.Active = false
	user.RefreshTokenID = refreshTokenID.String()
	usersRepo := &fakeUsersRepo{
		onSelectUser: user,
	}
	svc := NewService(usersRepo, "opa")

	_, err := svc.Refresh(t.Context(), validTokenForUser(t, svc, user.ID, refreshTokenID))
	if !errors.Is(err, ErrAuthFailed) && !strings.Contains(err.Error(), "deactivated") {
		t.Fatalf("expected auth error, got something else: %v", err)
	}
}

func TestRefreshWithNonExistentUser(t *testing.T) {

	dbError := errors.New("opa x2")

	usersRepo := &fakeUsersRepo{
		onSelectUser:    users.User{},
		errOnSelectUser: dbError,
	}
	svc := NewService(usersRepo, "opa")

	_, err := svc.Refresh(t.Context(), validTokenForUser(t, svc, uuid.New(), uuid.New()))
	if !errors.Is(err, ErrAuthFailed) && !strings.Contains(err.Error(), "opa x2") {
		t.Fatalf("expected auth error, got something else: %v", err)
	}
}

func TestRefreshSetsUserID(t *testing.T) {

	refreshTokenID := uuid.New()
	user := newUser()
	user.Active = true
	user.RefreshTokenID = refreshTokenID.String()
	usersRepo := &fakeUsersRepo{
		onSelectUser: user,
	}
	svc := NewService(usersRepo, "opa")

	pair, err := svc.Refresh(t.Context(), validTokenForUser(t, svc, user.ID, refreshTokenID))
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}

	if !tokenContainsUserID(t, svc, pair.Access, user.ID) {
		t.Errorf("access token has incorrect user id: got=%q, want=%q", pair.Access, user.ID)
	}

	if !tokenContainsUserID(t, svc, pair.Refresh, user.ID) {
		t.Errorf("refresh token has incorrect user id: got=%q, want=%q", pair.Access, user.ID)
	}
}

func TestRefreshSetsTokenDuration(t *testing.T) {

	refreshTokenID := uuid.New()
	user := newUser()
	user.Active = true
	user.RefreshTokenID = refreshTokenID.String()
	usersRepo := &fakeUsersRepo{
		onSelectUser: user,
	}
	svc := NewService(usersRepo, "opa")

	pair, err := svc.Refresh(t.Context(), validTokenForUser(t, svc, user.ID, refreshTokenID))
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}

	if got, want := readTokenDuration(t, svc, pair.Access), 15*time.Minute; got != want {
		t.Errorf("Access token has incorrect duration: got=%s, want=%s", got, want)
	}
	if got, want := readTokenDuration(t, svc, pair.Refresh), 7*24*time.Hour; got != want {
		t.Errorf("Refresh token has incorrect duration: got=%s, want=%s", got, want)
	}
}

func TestRefreshWithIncorrectTokenID(t *testing.T) {

	refreshTokenID := uuid.New()
	user := newUser()
	user.Active = false
	user.RefreshTokenID = refreshTokenID.String()
	usersRepo := &fakeUsersRepo{
		onSelectUser: user,
	}
	svc := NewService(usersRepo, "opa")

	_, err := svc.Refresh(t.Context(), validTokenForUser(t, svc, user.ID, uuid.New()))
	if !errors.Is(err, ErrAuthFailed) && !strings.Contains(err.Error(), "jti") {
		t.Fatalf("expected auth error, got something else: %v", err)
	}
}

func TestUserFromTokenWithValidToken(t *testing.T) {

	expectedUser := newUser()
	usersRepo := &fakeUsersRepo{
		onSelectUser: expectedUser,
	}
	svc := NewService(usersRepo, "opa")

	tokenString := validTokenForUser(t, svc, expectedUser.ID, uuid.New())

	loadedUser, _, err := svc.userFromToken(t.Context(), tokenString)
	if err != nil {
		t.Fatalf("expected nil, got error: %v", err)
	}

	if loadedUser != expectedUser {
		t.Errorf("users don't match")
	}

}

func TestUserFromTokenWithDbError(t *testing.T) {

	dbError := errors.New("")

	expectedUser := newUser()
	usersRepo := &fakeUsersRepo{
		onSelectUser:    users.User{},
		errOnSelectUser: dbError,
	}
	svc := NewService(usersRepo, "opa")

	tokenString := validTokenForUser(t, svc, expectedUser.ID, uuid.New())

	_, _, err := svc.userFromToken(t.Context(), tokenString)
	if !errors.Is(err, dbError) {
		t.Fatalf("expected db error, got: %v", err)
	}

}

func TestUserFromTokenWithInvalidToken(t *testing.T) {

	usersRepo := &fakeUsersRepo{
		onSelectUser: users.User{},
	}
	svc := NewService(usersRepo, "opa")

	_, _, err := svc.userFromToken(t.Context(), "opa")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token error, got: %v", err)
	}

}

func TestNewTokenPairSetsUserID(t *testing.T) {

	user := newUser()
	svc := NewService(&fakeUsersRepo{}, "opa")

	pair, err := svc.newTokenPair(t.Context(), user)
	if err != nil {
		t.Fatalf("error when creating new token pair: %v", err)
	}

	if !tokenContainsUserID(t, svc, pair.Access, user.ID) {
		t.Errorf("access token has incorrect user id: got=%q, want=%q", pair.Access, user.ID)
	}

	if !tokenContainsUserID(t, svc, pair.Refresh, user.ID) {
		t.Errorf("refresh token has incorrect user id: got=%q, want=%q", pair.Access, user.ID)
	}

}

func TestNewTokenPairSetsDuration(t *testing.T) {

	user := newUser()
	svc := NewService(&fakeUsersRepo{}, "opa")

	pair, err := svc.newTokenPair(t.Context(), user)
	if err != nil {
		t.Fatalf("error when creating new token pair: %v", err)
	}

	if got, want := readTokenDuration(t, svc, pair.Access), 15*time.Minute; got != want {
		t.Errorf("Access token has incorrect duration: got=%s, want=%s", got, want)
	}
	if got, want := readTokenDuration(t, svc, pair.Refresh), 7*24*time.Hour; got != want {
		t.Errorf("Refresh token has incorrect duration: got=%s, want=%s", got, want)
	}

}

func TestNewTokenPairUpdatesUser(t *testing.T) {

	user := newUser()
	usersRepo := &fakeUsersRepo{}
	svc := NewService(usersRepo, "opa")

	pair, err := svc.newTokenPair(t.Context(), user)
	if err != nil {
		t.Fatalf("error when creating new token pair: %v", err)
	}

	if len(usersRepo.updateCallArgs) != 1 {
		t.Errorf("Expected users repo to be called once, called %d times", len(usersRepo.updateCallArgs))
	}

	expectedRefreshTokenID := readTokenID(t, svc, pair.Refresh)
	updatedUser := usersRepo.updateCallArgs[0]

	if updatedUser.RefreshTokenID != expectedRefreshTokenID {
		t.Errorf("token ids do not match: user has %q, but token id is %q", updatedUser.RefreshTokenID, expectedRefreshTokenID)
	}

}

func TestNewToken(t *testing.T) {

	svc := NewService(&fakeUsersRepo{}, "opa")

	_, err := svc.newToken(t.Context(), uuid.New(), 15*time.Second, uuid.New())

	if err != nil {
		t.Errorf("expected nil, got error %q", err)
	}

}

func TestParseTokenWithValidTokenString(t *testing.T) {

	svc := NewService(&fakeUsersRepo{}, "opa")
	now := time.Now()
	expectedUserID := "some user id"
	expectedTokenID := "some token id"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims(expectedUserID, now, now.Add(15*time.Minute), expectedTokenID))
	tokenString, err := token.SignedString(svc.secretKey)
	if err != nil {
		t.Fatalf("unable to create token: %v", err)
	}

	parsedClaims, err := svc.parseToken(t.Context(), tokenString)
	if err != nil {
		t.Errorf("want nil, got error when parsing token string: %v", err)
	}

	if parsedClaims.ID != expectedTokenID {
		t.Errorf("parsed token id differs: got=%q, want=%q", parsedClaims.ID, expectedTokenID)
	}
	if parsedClaims.UserID != expectedUserID {
		t.Errorf("parsed user id differs: got=%q, want=%q", parsedClaims.UserID, expectedUserID)
	}

}

func TestParseTokenWithExpiredToken(t *testing.T) {

	svc := NewService(&fakeUsersRepo{}, "opa")
	now, err := time.Parse(time.RFC3339, "2024-10-10T14:00:00Z")
	if err != nil {
		t.Fatalf("unable to parse date: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims("not important", now, now.Add(15*time.Minute), "not important"))
	tokenString, err := token.SignedString(svc.secretKey)
	if err != nil {
		t.Fatalf("unable to create token: %v", err)
	}

	_, err = svc.parseToken(t.Context(), tokenString)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("want invalid token error, got %v", err)
	}

}

func TestParseTokenWithInvalidToken(t *testing.T) {

	svc := NewService(&fakeUsersRepo{}, "opa")
	_, err := svc.parseToken(t.Context(), "opa-opa")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("want invalid token error, got %v", err)
	}

}
