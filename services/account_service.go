package services

import (
	"zapmeow/models"
	"zapmeow/repositories"
)

type AccountService interface {
	CreateAccount(account *models.Account) error
	GetConnectedAccounts() ([]models.Account, error)
	GetAccountByInstanceID(instanceID string) (*models.Account, error)
	UpdateAccount(instanceID string, data map[string]interface{}) error
}

type accountService struct {
	accountRepo repositories.AccountRepository
}

func NewAccountService(accountRepo repositories.AccountRepository) *accountService {
	return &accountService{
		accountRepo: accountRepo,
	}
}

func (a *accountService) CreateAccount(account *models.Account) error {
	return a.accountRepo.CreateAccount(account)
}

func (a *accountService) GetConnectedAccounts() ([]models.Account, error) {
	return a.accountRepo.GetConnectedAccounts()
}

func (a *accountService) GetAccountByInstanceID(instanceID string) (*models.Account, error) {
	return a.accountRepo.GetAccountByInstanceID(instanceID)
}

func (a *accountService) UpdateAccount(instanceID string, data map[string]interface{}) error {
	return a.accountRepo.UpdateAccount(instanceID, data)
}
