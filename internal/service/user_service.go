package service

import (
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(r *repository.UserRepository) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) CreateUser(input model.CreateUserInput) (*model.User, error) {

	user := model.User{
		Name:  input.Name,
		Email: input.Email,
	}

	err := s.repo.CreateUser(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) GetAllUsers() ([]model.User, error) {
	return s.repo.FindAllUsers()
}
