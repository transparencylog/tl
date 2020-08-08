package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"go.transparencylog.net/tl/clientcache/badger"
)

const ServerAddr string = "beta-asset.transparencylog.net"
const ServerKey string = "log+3809a75e+ARmkoBH4C+/rbs9QomTtpLJQCkzfY171BfHZLEnmA/+e"

// ClientCache returns an initialized ClientCache using ServerAddr and ServerKey
func ClientCache() *badger.ClientCache {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tlDir := filepath.Join(home, ".config", "tl")
	err = os.MkdirAll(tlDir, 0700)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cacheFile := filepath.Join(tlDir, "tl.badger.db")

	// Initialize cache DB, if necessary
	cache := badger.NewClientCache(cacheFile, ServerAddr)
	_, err = cache.ReadConfig("key")
	if err != nil {
		if err := cache.WriteConfig("key", nil, []byte(ServerKey)); err != nil {
			log.Fatal(err)
		}
	}

	return cache
}
