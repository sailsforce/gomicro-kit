package mapper

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Product struct {
	Base
	Name         string `gorm:"unique" json:"name"`
	BusinessLine string `json:"business_line"`
	ProductLine  string `json:"product_line"`
	GUID         string `gorm:"column:guid" json:"guid"`
	DisplayName  string `json:"display_name"`
	Cloud        string `json:"cloud"`
}

type ProductCache struct {
	Map map[string]string
}

func MapProductGUID(data interface{}, logger logrus.FieldLogger, db *gorm.DB, cache *ProductCache) (interface{}, error) {
	fieldName, productValue := GetFieldByFieldTag("mapper", "product", data)
	if productValue == nil {
		return nil, errors.New("product value nil from GetFieldValueByFieldTag")
	}
	guidStr := findGUIDFromCache(fmt.Sprintf("%v", reflect.ValueOf(productValue)), logger, cache.Map)
	logger.Debug("field name: ", fieldName, " | value: ", productValue, " | guid str: ", guidStr)
	return updateFieldValueByName(data, fieldName, guidStr)
}

func (c *ProductCache) LoadProductCache(db *gorm.DB) error {
	var products []Product
	if err := db.Where("guid != ''").Find(&products).Error; err != nil {
		return err
	}
	c.Map = make(map[string]string)
	for _, v := range products {
		c.Map[v.Name] = v.GUID
	}
	return nil
}
