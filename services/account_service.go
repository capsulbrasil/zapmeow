package services

import (
	"fmt"
	"os"
	"path/filepath"
	"zapmeow/models"
	"zapmeow/repositories"
	"zapmeow/utils"
)

type AccountService interface {
	CreateAccount(account *models.Account) error
	GetConnectedAccounts() ([]models.Account, error)
	GetAccountByInstanceID(instanceID string) (*models.Account, error)
	UpdateAccount(instanceID string, data map[string]interface{}) error
	DeleteAccountMessages(instanceID string) error
}

type accountService struct {
	accountRepo    repositories.AccountRepository
	messageService MessageService
}

func NewAccountService(accountRepo repositories.AccountRepository, messageService MessageService) *accountService {
	return &accountService{
		accountRepo:    accountRepo,
		messageService: messageService,
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

func (a *accountService) DeleteAccountMessages(instanceID string) error {
	err := a.messageService.DeleteMessagesByInstanceID(instanceID)
	if err != nil {
		return err
	}
	return a.deleteAccountDirectory(instanceID)
}

func (a *accountService) deleteAccountDirectory(instanceID string) error {
	dirPath := utils.MakeAccountStoragePath(instanceID)
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			err = os.Remove(path)
			if err != nil {
				return err
			}
			fmt.Printf("File removed: %s\n", path)
		}
		return nil
	})
	return err
}
