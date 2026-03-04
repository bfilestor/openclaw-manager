package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrTokenMissing = errors.New("token missing")
	ErrTokenInvalid = errors.New("token invalid")
	ErrTokenExpired = errors.New("token expired")
	ErrTokenRevoked = errors.New("token revoked")
)

type BlacklistChecker interface {
	ExistsJTI(jti string) (bool, error)
}

type JWTService struct {
	Secret           []byte
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	BlacklistChecker BlacklistChecker
}

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (s *JWTService) SignAccessToken(userID, role string) (token string, jti string, err error) {
	jti = uuid.NewString()
	now := time.Now()
	c := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.AccessTokenTTL)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	token, err = t.SignedString(s.Secret)
	return
}

func (s *JWTService) VerifyAccessToken(tokenStr string) (*Claims, error) {
	if tokenStr == "" {
		return nil, ErrTokenMissing
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		return s.Secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	if !token.Valid {
		return nil, ErrTokenInvalid
	}
	if s.BlacklistChecker != nil {
		revoked, err := s.BlacklistChecker.ExistsJTI(claims.ID)
		if err != nil {
			return nil, err
		}
		if revoked {
			return nil, ErrTokenRevoked
		}
	}
	return claims, nil
}

func (s *JWTService) SignRefreshToken() (tokenID, rawToken string, err error) {
	tokenID = uuid.NewString()
	rawToken = uuid.NewString() + uuid.NewString()
	return
}
