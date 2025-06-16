package services

import (
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
)

type IAuthServicer interface {
	SignUp(user models.User) error
}

type AuthService struct {
	repo repositories.IAuthRepository
}

func NewAuthService(r repositories.IAuthRepository) IAuthServicer {
	return &AuthService{repo: r}
}

func (s *AuthService) SignUp(user models.User) error {
	//TODO
	return nil
}
