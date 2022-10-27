package mapper

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
)

type Base struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
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

func updateFieldValueByName(data interface{}, fieldName, newValue string) (interface{}, error) {
	reflect.ValueOf(data).Elem().FieldByName(fieldName).SetString(newValue)
	return data, nil

}
