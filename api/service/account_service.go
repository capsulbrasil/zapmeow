package service

import (
	"os"
	"path/filepath"
	"zapmeow/api/helper"
	"zapmeow/api/model"
	"zapmeow/api/repository"
)

type AccountService interface {
	CreateAccount(account *model.Account) error
	GetConnectedAccounts() ([]model.Account, error)
	GetAccountByInstanceID(instanceID string) (*model.Account, error)
	UpdateAccount(instanceID string, data map[string]interface{}) error
	DeleteAccountMessages(instanceID string) error
}

type accountService struct {
	accountRepo    repository.AccountRepository
	messageService MessageService
}

func NewAccountService(accountRepo repository.AccountRepository, messageService MessageService) *accountService {
	return &accountService{
		accountRepo:    accountRepo,
		messageService: messageService,
	}
}

func (a *accountService) CreateAccount(account *model.Account) error {
	return a.accountRepo.CreateAccount(account)
}

func (a *accountService) GetConnectedAccounts() ([]model.Account, error) {
	return a.accountRepo.GetConnectedAccounts()
}

func (a *accountService) GetAccountByInstanceID(instanceID string) (*model.Account, error) {
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
	dirPath := helper.MakeAccountStoragePath(instanceID)
	_, err := os.Stat(dirPath)
	if err != nil {
		return nil
	}

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			err = os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
