package repositories

import (
	"zapmeow/models"

	"gorm.io/gorm"
)

type AccountRepository interface {
	CreateAccount(account *models.Account) error
	GetConnectedAccounts() ([]models.Account, error)
	GetAccountByInstanceID(instanceID string) (*models.Account, error)
	UpdateAccount(instanceID string, data map[string]interface{}) error
}

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *accountRepository {
	return &accountRepository{db: db}
}

func (repo *accountRepository) CreateAccount(account *models.Account) error {
	return repo.db.Create(account).Error
}

func (repo *accountRepository) GetConnectedAccounts() ([]models.Account, error) {
	var accounts []models.Account
	repo.db.Where("status = ?", "CONNECTED").Find(&accounts)
	return accounts, nil
}

func (repo *accountRepository) GetAccountByInstanceID(instanceID string) (*models.Account, error) {
	var account models.Account
	result := repo.db.Where("instance_id = ?", instanceID).First(&account)
	if result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound {
			return nil, result.Error
		}
		return nil, nil
	}
	return &account, nil
}

func (repo *accountRepository) UpdateAccount(instanceID string, data map[string]interface{}) error {
	var account models.Account
	if result := repo.db.Where("instance_id = ?", instanceID).First(&account); result.Error != nil {
		return result.Error
	}

	if err := repo.db.Model(&account).Updates(data).Error; err != nil {
		return err
	}

	return nil
}
