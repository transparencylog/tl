package sqlite

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"go.transparencylog.net/btget/sumdb"
	_ "rsc.io/sqlite"
)

type ClientCache struct {
	sql        *sql.DB
	serverAddr string
}

func NewClientCache(cacheFile string, serverAddr string) *ClientCache {
	client := &ClientCache{serverAddr: serverAddr}

	sdb, err := sql.Open("sqlite3", cacheFile)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := sdb.Exec(`create table if not exists kv (k primary key, v)`); err != nil {
		log.Fatal(err)
	}
	client.sql = sdb
	return client
}

func (c *ClientCache) ReadRemote(path string) ([]byte, error) {
	resp, err := http.Get("https://" + c.serverAddr + path)
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

	data, err = sqlRead(c.sql, "config:"+file)
	if strings.HasSuffix(file, "/latest") && err == sql.ErrNoRows {
		return nil, nil
	}
	return data, err
}

func (c *ClientCache) WriteConfig(file string, old, new []byte) error {
	if old == nil {
		return sqlWrite(c.sql, "config:"+file, new)
	}
	return sqlSwap(c.sql, "config:"+file, old, new)
}

func (c *ClientCache) ReadCache(file string) ([]byte, error) {
	return sqlRead(c.sql, "file:"+file)
}

func (c *ClientCache) WriteCache(file string, data []byte) {
	sqlWrite(c.sql, "file:"+file, data)
}

func (c *ClientCache) Log(msg string) {
	log.Print(msg)
}

func (c *ClientCache) SecurityError(msg string) {
	log.Fatal(msg)
}

func sqlRead(db *sql.DB, key string) ([]byte, error) {
	var value []byte
	err := db.QueryRow(`select v from kv where k = ?`, key).Scan(&value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func sqlWrite(db *sql.DB, key string, value []byte) error {
	_, err := db.Exec(`insert or replace into kv (k, v) values (?, ?)`, key, value)
	return err
}

func sqlSwap(db *sql.DB, key string, old, value []byte) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var txOld []byte
	if err := tx.QueryRow(`select v from kv where k = ?`, key).Scan(&txOld); err != nil {
		return err
	}
	if !bytes.Equal(txOld, old) {
		return sumdb.ErrWriteConflict
	}
	if _, err := tx.Exec(`insert or replace into kv (k, v) values (?, ?)`, key, value); err != nil {
		return err
	}
	return tx.Commit()
}
