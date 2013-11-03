package main

import (
	// "bytes"
	// "encoding/json"
	"flag"
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
	// "io/ioutil"
	// "net/http"
	// "net/url"
	"os"
	// "strconv"
	"time"
)

type WoWItem struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

var (
	createBigDoc bool
	docSize      int
)

func init() {
	flag.BoolVar(&createBigDoc, "createbigdoc", false, "Create a big document with all item names")
	flag.IntVar(&docSize, "docsize", 10, "Document size")
	flag.Parse()
}

func main() {

	t0 := time.Now()
	wowb := Connect("wowitems")
	casb := Connect("castest")

	if createBigDoc {
		wowitems := make([]WoWItem, 0)
		fmt.Printf("Creating big composite document of all item names\n")

		_ = casb.Delete("wowitems")

		for i := 1000; i < 100000; i++ {
			var wowItem WoWItem
			key := fmt.Sprintf("item_%d", i)
			e := wowb.Get(key, &wowItem)

			if e != nil {
				continue
			}

			if i%100 == 0 {
				fmt.Printf("Parsing %d wowitems length is %d \n", i, len(wowitems))
			}

			if wowItem.Id > 0 {
				fmt.Printf("Found Id %d and Name %s\n", wowItem.Id, wowItem.Name)
				wowitems = append(wowitems, wowItem)
			}

			if len(wowitems) >= 10 {
				break
			}
		}

		fmt.Printf("Built out %s", len(wowitems))

		casb.Set("wowitems", -1, wowitems)
	}

	fmt.Printf("Getting CAS Value for 'wowitems'")

	meta, e := casb.Observe("wowitems")
	fmt.Printf("CAS==%d ERROR:%s", meta.Cas, e)

	fmt.Printf("Done! %s\n\n", time.Now().Sub(t0))
}

func Connect(n string) (bucket *couchbase.Bucket) {
	bucket, err := couchbase.GetBucket("http://localhost:8091/", "default", n)
	if err != nil {
		fmt.Printf("COUCHBASE FAILURE: %s\n", err)
		os.Exit(1)
	}
	return bucket
}
