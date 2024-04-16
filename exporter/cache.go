package exporter

import (
	"github.com/akrylysov/pogreb"
	"github.com/prometheus/common/log"
)

type Cache struct {
	db *pogreb.DB
}

func NewCache() *Cache {
	db, err := pogreb.Open("s3.cache", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return &Cache{
		db: db,
	}
}

func (c *Cache) Set(key, val string) {
	if err := c.db.Put([]byte(key), []byte(val)); err != nil {
		log.Fatal(err)
	}
}

func (c *Cache) Get(key string) {
	val, err := c.db.Get([]byte(key))
	if err != nil {
		log.Fatal(err)
	}
	log.Warnln(string(val))
}
