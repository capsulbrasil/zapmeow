package repositories

import (
	"errors"
	"zapmeow/models"

	"gorm.io/gorm"
)

type ProxyRepository interface {
	CreateProxy(proxy *models.Proxy) error
	UpdateProxy(id uint, data map[string]interface{}) error
	DeleteProxy(id uint) error
	GetProxy(id uint) (*models.Proxy, error)
	GetProxyWithHighestRanking() (*models.Proxy, error)
	GetProxyByIPPortScheme(ip, port, scheme string) (*models.Proxy, error)
}

type proxyRepository struct {
	db *gorm.DB
}

func NewProxyRepository(db *gorm.DB) *proxyRepository {
	return &proxyRepository{db: db}
}

func (repo *proxyRepository) CreateProxy(proxy *models.Proxy) error {
	return repo.db.Create(proxy).Error
}

func (repo *proxyRepository) UpdateProxy(id uint, data map[string]interface{}) error {
	var proxy models.Proxy
	if result := repo.db.Where("id = ?", id).First(&proxy); result.Error != nil {
		return result.Error
	}

	if err := repo.db.Model(&proxy).Updates(data).Error; err != nil {
		return err
	}

	return nil
}

func (repo *proxyRepository) GetProxy(id uint) (*models.Proxy, error) {
	var proxy models.Proxy
	result := repo.db.Where("id = ?", id).First(&proxy)
	if result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound {
			return nil, result.Error
		}
		return nil, nil
	}
	return &proxy, nil

}

func (repo *proxyRepository) GetProxyByIPPortScheme(ip, port, scheme string) (*models.Proxy, error) {
	var proxy models.Proxy
	err := repo.db.Where("ip = ? AND port = ? AND scheme = ?", ip, port, scheme).First(&proxy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &proxy, nil
}

func (repo *proxyRepository) GetProxyWithHighestRanking() (*models.Proxy, error) {
	var proxy models.Proxy
	if err := repo.db.Order("ranking DESC").First(&proxy).Error; err != nil {
		return nil, err
	}
	return &proxy, nil
}

func (repo *proxyRepository) DeleteProxy(id uint) error {
	if result := repo.db.Where("id = ?", id).Unscoped().Delete(&models.Proxy{}); result.Error != nil {
		return result.Error
	}
	return nil
}
