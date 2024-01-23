package repository

import (
	"zapmeow/api/model"
	"zapmeow/pkg/database"

	"gorm.io/gorm"
)

type AccountRepository interface {
	CreateAccount(account *model.Account) error
	GetConnectedAccounts() ([]model.Account, error)
	GetAccountByInstanceID(instanceID string) (*model.Account, error)
	UpdateAccount(instanceID string, data map[string]interface{}) error
}

type accountRepository struct {
	database database.Database
}

func NewAccountRepository(database database.Database) *accountRepository {
	return &accountRepository{database: database}
}

func (repo *accountRepository) CreateAccount(account *model.Account) error {
	return repo.database.Client().Create(account).Error
}

func (repo *accountRepository) GetConnectedAccounts() ([]model.Account, error) {
	var accounts []model.Account
	repo.database.Client().Where("status = ?", "CONNECTED").Find(&accounts)
	return accounts, nil
}

func (repo *accountRepository) GetAccountByInstanceID(instanceID string) (*model.Account, error) {
	var account model.Account
	result := repo.database.Client().Where("instance_id = ?", instanceID).First(&account)
	if result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound {
			return nil, result.Error
		}
		return nil, nil
	}
	return &account, nil
}

func (repo *accountRepository) UpdateAccount(instanceID string, data map[string]interface{}) error {
	var account model.Account
	if result := repo.database.Client().Where("instance_id = ?", instanceID).First(&account); result.Error != nil {
		return result.Error
	}

	if err := repo.database.Client().Model(&account).Updates(data).Error; err != nil {
		return err
	}

	return nil
}
