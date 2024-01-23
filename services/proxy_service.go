package services

import (
	"zapmeow/models"
	"zapmeow/repositories"
	"zapmeow/utils"
)

type ProxyService interface {
	CreateProxy(proxy *models.Proxy) error
	UpdateProxy(id uint, data map[string]interface{}) error
	DeleteProxy(id uint) error
	GetProxy(id uint) (*models.Proxy, error)
	GetProxyWithHighestRanking() (*models.Proxy, error)
}

type proxyService struct {
	proxyRepo repositories.ProxyRepository
}

func NewProxyService(proxyRepo repositories.ProxyRepository) *proxyService {
	return &proxyService{
		proxyRepo: proxyRepo,
	}
}

func (a *proxyService) GetProxyWithHighestRanking() (*models.Proxy, error) {
	return a.proxyRepo.GetProxyWithHighestRanking()
}

func (a *proxyService) CreateProxy(proxy *models.Proxy) error {
	return a.proxyRepo.CreateProxy(proxy)
}

func (a *proxyService) UpdateProxy(id uint, data map[string]interface{}) error {
	return a.proxyRepo.UpdateProxy(id, data)
}

func (a *proxyService) GetProxy(id uint) (*models.Proxy, error) {
	return a.proxyRepo.GetProxy(id)
}

func (a *proxyService) DeleteProxy(id uint) error {
	return a.proxyRepo.DeleteProxy(id)
}

func (a *proxyService) CreateProxyIfNotExists(proxy *models.Proxy) error {
	existingProxy, err := a.proxyRepo.GetProxyByIPPortScheme(proxy.Ip, proxy.Port, proxy.Scheme)
	if err != nil {
		return err
	}
	if existingProxy != nil {
		return nil
	}
	return a.CreateProxy(proxy)
}

func (a *proxyService) FromJSON(path string) error {
	proxys, err := utils.ReadProxys(path)
	if err != nil {
		return err
	}

	for _, p := range proxys.Proxys {
		err := a.CreateProxyIfNotExists(&models.Proxy{
			Scheme:  p.Scheme,
			Ip:      p.Ip,
			Port:    p.Port,
			Ranking: 10,
			Using:   0,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
