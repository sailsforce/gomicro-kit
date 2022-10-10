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

	})
})
