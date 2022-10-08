package config

import (
	"bytes"
	"database/sql"
	"log"
	"os"

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

type otherRuntimeVariables map[string]interface{}

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
	RV       otherRuntimeVariables
	Logger   *logrus.Logger
	// default to be used when setting up config.
	DB *gorm.DB
	// only used if more than one db is needed.
	DBList   []*gorm.DB
	HmacKeys *models.HmacKeys
}

func (c *MicroRestConfig) DefaultMicroConfig() {
	c.LoadNewRelicInfo()
	c.LoadServiceInfo()
	c.LoadLogger()
	c.LoadDatabases(c.Logger.Level, "DATABASE_URL")
}

func (c *MicroRestConfig) LoadNewRelicInfo() {
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
			log.Fatal("error setting up new relic logs: ", err)
		}
		c.NewRelic.App = relic
	}
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

func (c *MicroRestConfig) AddRV(key string, val interface{}) {
	c.RV[key] = val
}

func (c *MicroRestConfig) AddRVFromEnv(envar string) {
	c.RV[envar] = os.Getenv(envar)
}

func (c *MicroRestConfig) LoadRV(vars ...string) {
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

func (c *MicroRestConfig) LoadDatabases(logLvl logrus.Level, dburls ...string) {
	if os.Getenv("DATABASE_URL") != "" {
		if len(dburls) == 1 {
			// use DB var in config
			db, err := connectToDB(dburls[0], logLvl)
			if err != nil {
				log.Fatal("error connecting to db: ", err)
				return
			}
			c.DB = db
		} else {
			for _, v := range dburls {
				db, err := connectToDB(v, logLvl)
				if err != nil {
					log.Fatal("error connecting to db: ", err)
					return
				}
				// use DBList to populate all the databases given.
				c.DBList = append(c.DBList, db)
			}
		}
	}
}

func (c *MicroRestConfig) LoadHMACKeys() {
	// load in hmac keys
	var hmacKeys models.HmacKeys
	err := render.DecodeJSON(bytes.NewReader([]byte(os.Getenv("HMAC_SECRETS"))), &hmacKeys)
	if err != nil {
		log.Fatal("error pullin in hmac secret json obj: ", err)
		return
	}
	c.HmacKeys = &hmacKeys
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
