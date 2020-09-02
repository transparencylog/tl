package cat

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.transparencylog.com/tl/config"
	"go.transparencylog.com/tl/sumdb"
)

var CatCmd = &cobra.Command{
	Use:   "cat [URL]",
	Short: "Cat the contents of a URL only if the contents can be verified with the asset transparency log",

	Args: cobra.ExactArgs(1),

	Run: cat,
}

func cat(cmd *cobra.Command, args []string) {
	durl := args[0]

	u, err := url.Parse(durl)
	if err != nil {
		panic(err)
	}
	key := u.Host + u.Path

	cache := config.ClientCache()
	client := sumdb.NewClient(cache)

	// Step 1: Generate sha256sum of the file
	resp, err := http.Get(durl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	sum := sha256.Sum256(body)

	want := "h1:" + base64.StdEncoding.EncodeToString(sum[:])

	// Step 2: Download the tlog entry for the URL
	_, data, err := client.LookupOpts(key, sumdb.LookupOpts{Digest: want})
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		if line == want {
			break
		}
		if strings.HasPrefix(line, "h1:") {
			log.Fatalf("file digest %x != log digest %x", sum, line)
		}
	}

	b := bytes.NewBuffer(body)

	// Step 3: cat it out
	b.WriteTo(os.Stdout)
}
