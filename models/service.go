package models

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	"github.com/go-chi/render"
	"github.com/sailsforce/gomicro-kit/utils"
	"gorm.io/datatypes"
)

type Service struct {
	ID              int            `json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"update_at"`
	DeletedAt       time.Time      `json:"deleted_at"`
	ServiceName     string         `json:"service_name"`
	ServiceSummary  string         `json:"service_summary"`
	ServiceOnline   bool           `json:"service_online"`
	ServiceProtocol string         `json:"service_protocol"`
	ServiceVersion  string         `json:"service_version"`
	BaseURL         string         `json:"base_url"`
	Routes          datatypes.JSON `json:"routes"`
}

// Only used for documentation. Not used for database
type NewService struct {
	ServiceName     string                 `json:"service_name"`
	ServiceSummary  string                 `json:"service_summary"`
	ServiceOnline   bool                   `json:"service_online"`
	ServiceProtocol string                 `json:"service_protocol"`
	ServiceVersion  string                 `json:"service_version"`
	BaseURL         string                 `json:"base_url"`
	Routes          map[string]interface{} `json:"routes"`
}

type ServicePool struct {
	Services []*Service
	Current  uint64
}

func (s *Service) Print() {
	log.Printf("\nName: %v\nBaseURL: %v\nRoutes: %+v", s.ServiceName, s.BaseURL, s.Routes)
}

func (s *Service) RegisterAtGateway(gatewayUrl string) error {
	body, err := json.Marshal(s)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", gatewayUrl, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// add hmac header for validation
	var keyList HmacKeys
	err = render.DecodeJSON(bytes.NewReader([]byte(os.Getenv("HMAC_SECRETS"))), &keyList)
	if err != nil {
		return err
	}
	key := keyList.GetLatestKey()

	hmacByte := utils.CreateHmacHash(req, key)
	hmac64 := base64.StdEncoding.EncodeToString(hmacByte)
	// add to request headers
	req.Header.Add("X-HMAC-HASH", hmac64)

	c := http.DefaultClient
	resp, err := c.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("error registering service. Status: %v | err: %v", resp.StatusCode, err)
	}

	return nil
}

func (sp *ServicePool) Print() {
	for _, v := range sp.Services {
		v.Print()
	}
	log.Printf("current: %v\n", sp.Current)
}

func (sp *ServicePool) AddService(s *Service) {
	sp.Services = append(sp.Services, s)
}

func (sp *ServicePool) nextIndex() int {
	return int(atomic.AddUint64(&sp.Current, uint64(1)) % uint64(len(sp.Services)))
}

func (sp *ServicePool) MarkServiceStatus(serviceID int, alive bool) {
	for _, s := range sp.Services {
		if s.ID == serviceID {
			s.ServiceOnline = alive
			break
		}
	}
}

func (sp *ServicePool) GetNextPeer() *Service {
	next := sp.nextIndex()
	l := len(sp.Services) + next
	for i := next; i < l; i++ {
		idx := i % len(sp.Services)
		if sp.Services[idx].ServiceOnline {
			if i != next {
				atomic.StoreUint64(&sp.Current, uint64(idx))
			}
			return sp.Services[idx]
		}
	}
	return nil
}

func (sp *ServicePool) HealthCheck() {
	for _, s := range sp.Services {
		status := "up"
		alive, err := isServiceAlive(s)
		if err != nil {
			log.Printf("error: %v", err)
		}
		s.ServiceOnline = alive
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", s.BaseURL, status)
	}
}

func isServiceAlive(s *Service) (bool, error) {
	var serviceRoutes map[string]string
	err := json.Unmarshal(s.Routes, &serviceRoutes)
	if err != nil {
		return false, err
	}
	heatlhRoute, ok := serviceRoutes["health"]
	if !ok {
		return false, errors.New("no health route")
	}
	serviceURL, err := url.Parse(fmt.Sprintf("%s://%s/%s%s", s.ServiceProtocol, s.BaseURL, s.ServiceVersion, heatlhRoute))
	if err != nil {
		return false, err
	}

	timeout := 5 * time.Second
	log.Printf("dial: %v", serviceURL.Host)
	conn, err := net.DialTimeout("tcp", serviceURL.Host+":80", timeout)
	if err != nil {
		return false, nil
	}
	_ = conn.Close()
	return true, nil
}
