package mapper

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Base struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
}

type Product struct {
	Base
	Name         string `gorm:"unique" json:"name"`
	BusinessLine string `json:"business_line"`
	ProductLine  string `json:"product_line"`
	GUID         string `gorm:"column:guid" json:"guid"`
	DisplayName  string `json:"display_name"`
	Cloud        string `json:"cloud"`
}

func MapProductGUID(data interface{}, logger logrus.FieldLogger, db *gorm.DB) (interface{}, error) {
	productCache := make(map[string]string)
	loadProductCache(productCache, db)

	fieldName, productValue := GetFieldByFieldTag("mapper", "product", data)
	if productValue == nil {
		return nil, errors.New("product value nil from GetFieldValueByFieldTag")
	}
	guidStr := findGUIDFromCache(fmt.Sprintf("%v", reflect.ValueOf(productValue)), logger, productCache)
	logger.Debug("field name: ", fieldName, " | value: ", productValue, " | guid str: ", guidStr)
	return updateFieldValueByName(data, fieldName, guidStr)
}

func updateFieldValueByName(data interface{}, fieldName, newValue string) (interface{}, error) {
	reflect.ValueOf(data).Elem().FieldByName(fieldName).SetString(newValue)
	return data, nil

}

func findGUIDFromCache(value string, logger logrus.FieldLogger, cache map[string]string) string {
	var res string
	parts := strings.Split(value, ";")
	for _, v := range parts {
		if guid, ok := cache[v]; ok {
			logger.Debug("found guid in cache")
			res = fmt.Sprintf("%s;%s", res, guid)
		} else {
			logger.Debug("guid not found for: ", v)
		}
	}
	return trimFirstRune(res)
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func loadProductCache(cache map[string]string, db *gorm.DB) error {
	var products []Product
	if err := db.Where("guid != ''").Find(&products).Error; err != nil {
		return err
	}
	for _, v := range products {
		cache[v.Name] = v.GUID
	}
	return nil
}

func GetFieldByFieldTag(jsonTagKey, jsonTagValue string, s interface{}) (string, interface{}) {
	rt := reflect.TypeOf(s).Elem()
	if rt.Kind() != reflect.Struct {
		return "", nil
	}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		v := strings.Split(rt.Field(i).Tag.Get(jsonTagKey), ",")[0]
		if v == jsonTagValue {
			r := reflect.ValueOf(s)
			field := reflect.Indirect(r).FieldByName(f.Name)
			return f.Name, field
		}
	}

	return "", nil
}
