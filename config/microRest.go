package config

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/render"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sailsforce/gomicro-kit/models"
	"github.com/sailsforce/gomicro-kit/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type newRelicInfo struct {
	AppName     string
	License     string
	DisplayName string
	App         *newrelic.Application
}

type serviceInfo struct {
	Name       string
	Summary    string
	Protocal   string
	Version    string
	BaseURL    string
	Routes     string
	GatewayURL string
}

type MicroRestConfig struct {
	NewRelic newRelicInfo
	Service  serviceInfo
	RV       map[string]interface{}
	Logger   *logrus.Logger
	// default to be used when setting up config.
	DB *gorm.DB
	// only used if more than one db is needed.
	DBList   []*gorm.DB
	HmacKeys *models.HmacKeys
}

func (c *MicroRestConfig) DefaultMicroConfig() error {
	if err := c.LoadNewRelicInfo(); err != nil {
		return err
	}
	c.LoadServiceInfo()
	c.LoadLogger()
	if err := c.LoadDatabases(c.Logger.Level, "DATABASE_URL"); err != nil {
		return err
	}
	return nil
}

func (c *MicroRestConfig) LoadNewRelicInfo() error {
	c.NewRelic = newRelicInfo{
		os.Getenv("NEW_RELIC_APP_NAME"),
		os.Getenv("NEW_RELIC_LICENSE"),
		os.Getenv("NEW_RELIC_DISPLAY_NAME"),
		nil,
	}
	// check to verify all variables are there, then make connection.
	if c.NewRelic.AppName != "" && c.NewRelic.License != "" && c.NewRelic.DisplayName != "" {
		relic, err := newrelic.NewApplication(
			newrelic.ConfigAppName(c.NewRelic.AppName),
			newrelic.ConfigLicense(c.NewRelic.License),
			newrelic.ConfigDistributedTracerEnabled(true),
			func(cfg *newrelic.Config) {
				cfg.ErrorCollector.RecordPanics = true
				cfg.HostDisplayName = c.NewRelic.DisplayName
			},
		)
		if err != nil {
			return fmt.Errorf("error creating NewRelic App: %v", err)
		}
		c.NewRelic.App = relic
	}
	return nil
}

func (c *MicroRestConfig) LoadServiceInfo() {
	c.Service = serviceInfo{
		os.Getenv("SERVICE_NAME"),
		os.Getenv("SERVICE_SUMMARY"),
		os.Getenv("SERVICE_PROTOCOL"),
		os.Getenv("SERVICE_VERSION"),
		os.Getenv("SERVICE_BASE_URL"),
		os.Getenv("SERVICE_ROUTES"),
		os.Getenv("GATEWAY_URL"),
	}
}

func (c *MicroRestConfig) RegisterAtGateway() error {
	c.Logger.Info("registering service...")
	var routesJson map[string]interface{}
	err := json.Unmarshal([]byte(c.Service.Routes), &routesJson)
	if err != nil {
		return fmt.Errorf("%s %v", "error parsing routes json: ", err)
	}
	routesBytes, err := json.Marshal(routesJson)
	if err != nil {
		return fmt.Errorf("%s %v", "error marshalling routes json: ", err)
	}
	service := models.Service{
		ServiceName:     c.Service.Name,
		ServiceSummary:  c.Service.Summary,
		ServiceOnline:   true,
		ServiceProtocol: c.Service.Protocal,
		ServiceVersion:  c.Service.Version,
		BaseURL:         c.Service.BaseURL,
		Routes:          routesBytes,
	}

	err = service.RegisterAtGateway(c.Service.GatewayURL)
	if err != nil {
		if strings.Contains(err.Error(), "409") {
			c.Logger.Info("service already registered.")
		} else {
			return fmt.Errorf("%s %v", "error registering service: ", err)
		}
	}

	c.Logger.Info("register complete at: ", c.Service.GatewayURL)
	return nil
}

func (c *MicroRestConfig) AddRV(key string, val interface{}) {
	if c.RV == nil {
		c.RV = make(map[string]interface{})
	}
	c.RV[key] = val
}

func (c *MicroRestConfig) AddRVFromEnv(envar string) {
	if c.RV == nil {
		c.RV = make(map[string]interface{})
	}
	c.RV[envar] = os.Getenv(envar)
}

func (c *MicroRestConfig) LoadRV(vars ...string) {
	if c.RV == nil {
		c.RV = make(map[string]interface{})
	}
	for _, v := range vars {
		c.RV[v] = os.Getenv(v)
	}
}

func (c *MicroRestConfig) LoadLogger() {
	newLog := logrus.New()
	newLog.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	newLog.SetOutput(os.Stdout)
	logLvl, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		newLog.Info("using default log level")
		logLvl = 4
	}
	newLog.SetLevel(logLvl)
	c.Logger = newLog
}

func (c *MicroRestConfig) LoadMockDatabase() {
	db, _, _ := sqlmock.New()
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.LogLevel(4)),
	})
	if err != nil {
		log.Fatal("error loading mock database: ", err)
		return
	}
	c.DB = gormDB
}

func (c *MicroRestConfig) LoadDatabases(logLvl logrus.Level, dburls ...string) error {
	if os.Getenv("DATABASE_URL") != "" {
		if len(dburls) == 1 {
			// use DB var in config
			db, err := connectToDB(os.Getenv(dburls[0]), logLvl)
			if err != nil {
				return fmt.Errorf("error connecting to db: %v", err)
			}
			c.DB = db
		} else {
			for _, v := range dburls {
				db, err := connectToDB(os.Getenv(v), logLvl)
				if err != nil {
					return fmt.Errorf("error connecting to db: %v", err)
				}
				// use DBList to populate all the databases given.
				c.DBList = append(c.DBList, db)
			}
		}
	}
	return nil
}

func (c *MicroRestConfig) LoadHMACKeys() error {
	// load in hmac keys
	var hmacKeys models.HmacKeys
	err := render.DecodeJSON(bytes.NewReader([]byte(os.Getenv("HMAC_SECRETS"))), &hmacKeys)
	if err != nil {
		return fmt.Errorf("error pullin in hmac secret json obj: %v", err)
	}
	c.HmacKeys = &hmacKeys
	return nil
}

func connectToDB(dburl string, logLvl logrus.Level) (*gorm.DB, error) {
	database, err := sql.Open("postgres", utils.GetDSN(dburl))
	if err != nil {
		return nil, err
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.LogLevel(logLvl)),
	})

	if err != nil {
		return nil, err
	}

	return gormDB, nil
}
