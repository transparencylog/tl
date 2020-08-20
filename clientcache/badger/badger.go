package badger

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	badger "github.com/dgraph-io/badger/v2"

	"go.transparencylog.com/tl/sumdb"
)

var ErrNoKey = errors.New("key not set")

type ClientCache struct {
	cacheFile  string
	serverURL  string
	bdbOptions badger.Options
}

func NewClientCache(cacheFile string, serverURL string) *ClientCache {
	client := &ClientCache{
		cacheFile:  cacheFile,
		serverURL:  serverURL,
		bdbOptions: badger.DefaultOptions(cacheFile).WithLogger(nil),
	}

	return client
}

func (c *ClientCache) ReadRemote(path string, query string) ([]byte, error) {
	u, err := url.Parse(c.serverURL)
	if err != nil {
		log.Fatalf("ReadRemote: %v", err)
	}
	u.Path = path
	u.RawQuery = query

	resp, err := http.Get(u.String())
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
	bdb, err := badger.Open(c.bdbOptions)
	if err != nil {
		return nil, err
	}
	defer bdb.Close()

	var value []byte

	err = bdb.View(func(tx *badger.Txn) error {
		item, err := tx.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			return ErrNoKey
		}
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			value = append(value, val...)
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return value, err
}

func (c *ClientCache) bdWrite(key string, value []byte) error {
	bdb, err := badger.Open(c.bdbOptions)
	if err != nil {
		return err
	}
	defer bdb.Close()

	err = bdb.Update(func(tx *badger.Txn) error {
		return tx.Set([]byte(key), value)
	})

	return err
}

func (c *ClientCache) bdSwap(key string, old, value []byte) error {
	bdb, err := badger.Open(c.bdbOptions)
	if err != nil {
		return err
	}
	defer bdb.Close()

	err = bdb.Update(func(tx *badger.Txn) error {
		var txOld []byte
		itemOld, err := tx.Get([]byte(key))
		if err != nil {
			return err
		}

		err = itemOld.Value(func(val []byte) error {
			txOld = append(txOld, val...)
			return nil
		})
		if err != nil {
			return err
		}

		if !bytes.Equal(txOld, old) {
			return sumdb.ErrWriteConflict
		}

		if err := tx.Set([]byte(key), value); err != nil {
			return err
		}
		return nil
	})

	return err
}
