package config

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config Tests", func() {

	c := MicroRestConfig{}

	AfterEach(func() {
		os.Clearenv()
		c = MicroRestConfig{}
	})

	Describe("LoadNewRelicInfo", func() {
		license40 := "1234567890123456789012345678901234567890"
		BeforeEach(func() {
			os.Setenv("NEW_RELIC_APP_NAME", "test app name")
			os.Setenv("NEW_RELIC_DISPLAY_NAME", "test display name")
		})
		AfterEach(func() {
			os.Clearenv()
		})
		Context("License is len 40", func() {
			It("should init NewRelic struct", func() {
				os.Setenv("NEW_RELIC_LICENSE", license40)
				err := c.LoadNewRelicInfo()
				// validate values
				Expect(err).To(BeNil())
				Expect(c.NewRelic.AppName).To(Equal("test app name"))
				Expect(c.NewRelic.License).To(Equal(license40))
				Expect(c.NewRelic.DisplayName).To(Equal("test display name"))
			})
		})
		Context("License is NOT len 40", func() {
			It("should fail when creating NewRelic.App", func() {
				os.Setenv("NEW_RELIC_LICENSE", "101")
				err := c.LoadNewRelicInfo()
				// validate values
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("error creating NewRelic App:"))
			})
		})
		Context("Not all the env vars are set", func() {
			It("should load in env vars but NOT init App", func() {
				err := c.LoadNewRelicInfo()
				// validate values
				Expect(err).To(BeNil())
				Expect(c.NewRelic.AppName).To(Equal("test app name"))
				Expect(c.NewRelic.DisplayName).To(Equal("test display name"))
			})
		})
	})

	Describe("LoadServiceInfo", func() {
		BeforeEach(func() {
			os.Setenv("SERVICE_NAME", "test service name")
			os.Setenv("SERVICE_SUMMARY", "test service summary")
			os.Setenv("SERVICE_PROTOCOL", "http")
			os.Setenv("SERVICE_VERSION", "api")
			os.Setenv("SERVICE_BASE_URL", "localhost:8080")
			os.Setenv("SERVICE_ROUTES", "{\"foo\": \"/bar\", \"health\": \"/heartbeat\"}")
			os.Setenv("GATEWAY_URL", "localhost:8880/register")
		})
		AfterEach(func() {
			os.Clearenv()
		})
		Context("All env vars set", func() {
			It("should load in env vars", func() {
				c.LoadServiceInfo()
				// validate values
				Expect(c.Service.Name).To(Equal("test service name"))
				Expect(c.Service.Summary).To(Equal("test service summary"))
				Expect(c.Service.Protocal).To(Equal("http"))
				Expect(c.Service.Version).To(Equal("api"))
				Expect(c.Service.BaseURL).To(Equal("localhost:8080"))
				Expect(c.Service.Routes).To(Equal("{\"foo\": \"/bar\", \"health\": \"/heartbeat\"}"))
				Expect(c.Service.GatewayURL).To(Equal("localhost:8880/register"))
			})
		})
		Describe("Register at Gateway", func() {
			Context("valid json, no gateway", func() {
				It("should fail since no gateway is running", func() {
					c.LoadLogger()
					c.LoadServiceInfo()
					err := c.RegisterAtGateway()
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("error registering service:"))
				})
			})
			Context("invalid json", func() {
				It("should fail unmarshalling json", func() {
					c.LoadLogger()
					c.LoadServiceInfo()
					c.Service.Routes = "{foo: \"/bar\", \"health\": \"/heartbeat\"}"
					err := c.RegisterAtGateway()
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("error parsing routes json:"))
				})
			})
		})
	})

	Describe("RuntimeVariables", func() {
		It("should add new key value to RV", func() {
			Expect(c.RV).To(BeNil())
			c.AddRV("foo", "bar")
			Expect(c.RV["foo"]).ToNot(BeNil())
			Expect(c.RV["foo"]).To(Equal("bar"))
		})
		It("should add rv from env var", func() {
			tempEnv := "TEMP_ENV"
			os.Setenv(tempEnv, "test-value")
			Expect(c.RV).To(BeNil())
			c.AddRVFromEnv(tempEnv)
			Expect(c.RV[tempEnv]).ToNot(BeNil())
			Expect(c.RV[tempEnv]).To(Equal("test-value"))
		})
		It("should load in many RVs from env vars", func() {
			envOne := "ENV_ONE"
			envTwo := "ENV_TWO"
			envThree := "ENV_THREE"
			os.Setenv(envOne, "one")
			os.Setenv(envTwo, "two")
			os.Setenv(envThree, "three")
			Expect(c.RV).To(BeNil())
			c.LoadRV(envOne, envTwo, envThree)
			Expect(c.RV[envOne]).ToNot(BeNil())
			Expect(c.RV[envTwo]).ToNot(BeNil())
			Expect(c.RV[envThree]).ToNot(BeNil())
			Expect(c.RV[envOne]).To(Equal("one"))
			Expect(c.RV[envTwo]).To(Equal("two"))
			Expect(c.RV[envThree]).To(Equal("three"))
		})
	})

	Describe("LoadLogger", func() {
		It("should use default log lvl", func() {
			Expect(c.Logger).To(BeNil())
			c.LoadLogger()
			Expect(c.Logger.Level.String()).To(Equal("info"))
		})
		It("should use debug log lvl", func() {
			os.Setenv("LOG_LEVEL", "debug")
			Expect(c.Logger).To(BeNil())
			c.LoadLogger()
			Expect(c.Logger.Level.String()).To(Equal("debug"))
		})
		It("should use error log lvl", func() {
			os.Setenv("LOG_LEVEL", "error")
			Expect(c.Logger).To(BeNil())
			c.LoadLogger()
			Expect(c.Logger.Level.String()).To(Equal("error"))
		})
	})

	Describe("LoadMockDatabase", func() {
		It("should init DB with sqlMock", func() {
			c.LoadLogger()
			Expect(c.DB).To(BeNil())
			c.LoadMockDatabase()
			Expect(c.DB).ToNot(BeNil())
		})
	})

	Describe("LoadDatabases", func() {
		It("should return error w/ one db url", func() {
			os.Setenv("DATABASE_URL", "postgres://postgres:admin@host.docker.internal:5437/postgres")
			c.LoadLogger()
			Expect(c.DB).To(BeNil())
			err := c.LoadDatabases(c.Logger.Level, "DATABASE_URL")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("error connecting to db:"))
		})
		It("should return error w/ many db urls", func() {
			os.Setenv("DATABASE_URL", "postgres://postgres:admin@host.docker.internal:5437/postgres")
			os.Setenv("ANOTHER_URL", "postgres://postgres:admin@host.docker.internal:5438/another")
			os.Setenv("THREE_DBS", "postgres://postgres:admin@host.docker.internal:5439/three")
			c.LoadLogger()
			Expect(c.DB).To(BeNil())
			err := c.LoadDatabases(c.Logger.Level, "DATABASE_URL", "ANOTHER_URL", "THREE_DBS")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("error connecting to db:"))
		})
	})

	Describe("LoadHMACKeys", func() {
		It("should pass and load object", func() {
			os.Setenv("HMAC_SECRETS", "{\"name\": \"Hmac keys\", \"keys\": [{\"created\": \"2021-10-12T18:00:42Z\", \"value\": \"supersecretkeyvalue\"},{\"created\": \"2020-10-13T18:00:42Z\", \"value\": \"anothersupersecretvalue\"}]}")
			err := c.LoadHMACKeys()
			Expect(err).To(BeNil())
			Expect(c.HmacKeys.Name).To(Equal("Hmac keys"))
			Expect(c.HmacKeys.Keys[0].Value).To(Equal("supersecretkeyvalue"))
			Expect(c.HmacKeys.Keys[1].Value).To(Equal("anothersupersecretvalue"))
		})
		It("should error", func() {
			os.Setenv("HMAC_SECRETS", "{name: \"Hmac keys\", \"keys\": [{\"created\": \"2021-10-12T18:00:42Z\", \"value\": \"supersecretkeyvalue\"},{\"created\": \"2020-10-13T18:00:42Z\", \"value\": \"anothersupersecretvalue\"}]}")
			err := c.LoadHMACKeys()
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("DefaultMicroConfig", func() {
		err := c.DefaultMicroConfig()
		Expect(err).To(BeNil())
		Expect(c.Logger.Level.String()).To(Equal("info"))
	})
})
