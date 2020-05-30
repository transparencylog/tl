package bbolt

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	bolt "go.etcd.io/bbolt"
	"go.transparencylog.net/btget/sumdb"
)

var bucket = []byte("btget")
var ErrNoKey = errors.New("key not set")

type ClientCache struct {
	cacheFile  string
	serverAddr string
}

func NewClientCache(cacheFile string, serverAddr string) *ClientCache {
	client := &ClientCache{cacheFile: cacheFile, serverAddr: serverAddr}

	bdb, err := bolt.Open(cacheFile, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer bdb.Close()

	err = bdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func (c *ClientCache) ReadRemote(path string) ([]byte, error) {
	resp, err := http.Get("https://" + c.serverAddr + path)

	println("https://" + c.serverAddr + path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http get: %v", resp.Status)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *ClientCache) ReadConfig(file string) (data []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("read config %s: %v", file, err)
		}
	}()

	data, err = c.bdRead("config:" + file)
	if strings.HasSuffix(file, "/latest") && err == ErrNoKey {
		return nil, nil
	}
	return data, err
}

func (c *ClientCache) WriteConfig(file string, old, new []byte) error {
	if old == nil {
		return c.bdWrite("config:"+file, new)
	}
	return c.bdSwap("config:"+file, old, new)
}

func (c *ClientCache) ReadCache(file string) ([]byte, error) {
	return c.bdRead("file:" + file)
}

func (c *ClientCache) WriteCache(file string, data []byte) {
	c.bdWrite("file:"+file, data)
}

func (c *ClientCache) Log(msg string) {
	log.Print(msg)
}

func (c *ClientCache) SecurityError(msg string) {
	log.Fatal(msg)
}

func (c *ClientCache) bdRead(key string) ([]byte, error) {
	bdb, err := bolt.Open(c.cacheFile, 0666, nil)
	if err != nil {
		return nil, err
	}
	defer bdb.Close()

	var value []byte

	err = bdb.View(func(tx *bolt.Tx) error {
		value = tx.Bucket(bucket).Get([]byte(key))
		if value == nil {
			return ErrNoKey
		}
		return nil
	})

	return value, err
}

func (c *ClientCache) bdWrite(key string, value []byte) error {
	bdb, err := bolt.Open(c.cacheFile, 0666, nil)
	if err != nil {
		return err
	}
	defer bdb.Close()

	err = bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if err := b.Put([]byte(key), value); err != nil {
			return err
		}
		return nil
	})

	return err
}

func (c *ClientCache) bdSwap(key string, old, value []byte) error {
	bdb, err := bolt.Open(c.cacheFile, 0666, nil)
	if err != nil {
		return err
	}
	defer bdb.Close()

	err = bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)

		txOld := b.Get([]byte(key))
		if !bytes.Equal(txOld, old) {
			return sumdb.ErrWriteConflict
		}

		if err := b.Put([]byte(key), value); err != nil {
			return err
		}
		return nil
	})

	return err
}
