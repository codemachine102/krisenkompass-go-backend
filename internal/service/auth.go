package service

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/config"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/repository"
	pass "bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/password"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/randstring"
	"github.com/dgrijalva/jwt-go"
)

type AuthService struct {
	userRepo             repository.User
	organizationUserRepo repository.OrganizationUser
}

func NewAuthService(userRepo repository.User, organizationUserRepo repository.OrganizationUser) *AuthService {
	return &AuthService{userRepo: userRepo, organizationUserRepo: organizationUserRepo}
}

func (s *AuthService) CreateNewSession(userID int64, userType string) (domain.Tokens, error) {
	tokens, err := s.CreateTokens(userID, userType)
	if err != nil {
		return tokens, err
	}

	// Save session in DB.
	err = s.userRepo.CreateRefreshSession(userID, tokens.RefreshToken, "", "", time.Now().Add(time.Hour*24*60))
	if err != nil {
		return domain.Tokens{}, err
	}

	return tokens, nil
}

func (s *AuthService) CreateTokens(userID int64, userType string) (domain.Tokens, error) {
	userInfo, err := s.organizationUserRepo.GetByUserID(userID)
	if err != nil {
		return domain.Tokens{}, err
	}
	roles := make(map[int64]string)
	for _, info := range userInfo {
		roles[info.OrganizationID] = info.Role
	}

	claims := domain.Claims{
		UserID: userID,
		Roles:  roles,
		Type:   userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(config.Get().App.Key))
	if err != nil {
		return domain.Tokens{}, err
	}

	refreshToken := base64.RawURLEncoding.EncodeToString([]byte(randstring.RandAlphanumString(128)))

	return domain.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Login(email, password string) (domain.Tokens, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Tokens{}, errors.New("user not found")
		}
		return domain.Tokens{}, err
	}

	// Check if account is verified.
	verification, err := s.userRepo.GetVerification(user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Tokens{}, errors.New("user not verified")
		}
		return domain.Tokens{}, err
	}
	if verification.Status == "not_verified" {
		return domain.Tokens{}, errors.New("user not verified")
	}

	err = pass.CheckPassword(password, user.Password.String)
	if err != nil {
		return domain.Tokens{}, err
	}

	return s.CreateNewSession(user.ID, user.Type)
}

func (s *AuthService) RefreshToken(refreshToken string) (domain.Tokens, error) {
	session, err := s.userRepo.GetRefreshSession(refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Tokens{}, errors.New("invalid token")
		}
		return domain.Tokens{}, err
	}

	// TODO: Check if token expired.

	user, err := s.userRepo.GetByID(session.UserID)
	if err != nil {
		return domain.Tokens{}, err
	}

	tokens, err := s.CreateTokens(user.ID, user.Type)
	if err != nil {
		return tokens, err
	}

	err = s.userRepo.UpdateRefreshSession(session.ID, tokens.RefreshToken, "", "", time.Now().Add(time.Hour*24*60))

	return tokens, err
}

func (s *AuthService) Verify(token, password string) (domain.Tokens, error) {
	user, err := s.userRepo.GetByVerificationToken(token)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Tokens{}, errors.New("access denied")

		}
		return domain.Tokens{}, err
	}

	verification, err := s.userRepo.GetVerification(user.ID)
	if err != nil {
		return domain.Tokens{}, err
	}

	if verification.Status == "verified" {
		return domain.Tokens{}, errors.New("access denied")
	}

	hashedPassword, err := pass.HashPassword(password)
	if err != nil {
		return domain.Tokens{}, err
	}

	err = s.userRepo.UpdatePassword(user.ID, hashedPassword)
	if err != nil {
		return domain.Tokens{}, err
	}

	err = s.userRepo.UpdateVerificationByToken(token, "verified")
	if err != nil {
		return domain.Tokens{}, err
	}

	return s.CreateNewSession(user.ID, user.Type)
}
