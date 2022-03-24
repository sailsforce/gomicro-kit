package models

import (
	"sort"
	"time"
)

type HmacKeys struct {
	Name string `json:"name"`
	Keys []Key  `json:"keys"`
}

type Key struct {
	Created time.Time `json:"created"`
	Value   string    `json:"value"`
}

func (hk *HmacKeys) GetLatestKey() string {
	sort.Slice(hk.Keys, func(i, j int) bool {
		return hk.Keys[i].Created.After(hk.Keys[j].Created)
	})

	return hk.Keys[0].Value
}
