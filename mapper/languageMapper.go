package mapper

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Language struct {
	// ID, CreatedAt, UpdatedAt, DeletedAt
	Base
	// Language Specific fields
	Name   string `gorm:"unique" json:"name"`
	GUID   string `gorm:"column:guid" json:"guid"`
	MCName string `gorm:"column:mc_name" json:"mc_name"`
}

type LanguageCache struct {
	Map map[string]string
}

func MapLanguageGUID(data interface{}, logger logrus.FieldLogger, db *gorm.DB, cache *LanguageCache) (interface{}, error) {
	fieldName, languageValue := GetFieldByFieldTag("mapper", "language", data)
	if languageValue == nil {
		return nil, errors.New("language value nil from GetFieldValueByFieldTag")
	}
	guidStr := findGUIDFromCache(fmt.Sprintf("%v", reflect.ValueOf(languageValue)), logger, cache.Map)
	logger.Debug("field name: ", fieldName, " | value: ", languageValue, " | guid str: ", guidStr)
	return updateFieldValueByName(data, fieldName, guidStr)
}

func (c *LanguageCache) LoadLanguageCache(db *gorm.DB) error {
	var languages []Language
	if err := db.Where("guid != ''").Find(&languages).Error; err != nil {
		return err
	}
	c.Map = make(map[string]string)
	for _, v := range languages {
		c.Map[v.Name] = v.GUID
	}

	return nil
}
