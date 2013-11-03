package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var start_id int //= 10000
var end_id int   //= 10500 //80000

type DocWithId struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Type        string `json:"type"`
}

type DocJustId struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func init() {
	flag.IntVar(&start_id, "startid", 10000, "Starting ID")
	flag.IntVar(&end_id, "endid", 20000, "Ending ID")
	flag.Parse()
}

func main() {

	fmt.Println("END.  This file will produce keys that can be parsed by the Couchbase-ElasticSearch Transport")
	os.Exit(2)

	fmt.Printf("Starting Import from %d to %d\n", start_id, end_id)
	t0 := time.Now()
	bucket := Connect()

	for i := start_id; i <= end_id; i++ {

		if i%50 == 0 {
			fmt.Printf("Importing %d\n", i)
		}

		importDoc(i, bucket)
	}

	fmt.Printf("Done!  %d items in %s\n\n", end_id-start_id, time.Now().Sub(t0))
}

func getDoc(i int) (bodyBytes []byte) {
	urlStr := "http://eu.battle.net/api/wow/item/" + strconv.Itoa(i)
	u, _ := url.Parse(urlStr)
	req := &http.Request{
		Method: "GET",
		URL:    u,
	}

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	bodyBytes, _ = ioutil.ReadAll(resp.Body)
	return
}

func importDoc(id int, bucket *couchbase.Bucket) {
	body := getDoc(id)

	dec := json.NewDecoder(bytes.NewReader(body))
	var d DocWithId
	dec.Decode(&d)
	if d.Status == "nok" || d.Name == "" {
		return
	}

	d.Type = "wowitem"

	d2 := &DocJustId{Id: d.Id, Name: d.Name}

	fmt.Printf("Retreived Doc %d (%s)\n", d.Id, d.Name)
	// b, _ := json.Marshal(d2)
	// _, _ = json.Marshal(d2)
	key := "item_" + strconv.Itoa(id)
	fmt.Printf("Key is %s (%#x)\t", key, key)
	// fmt.Printf("JSON Data == '%s'", string(b))
	bucket.Set("test", -1, d2)
	// bucket.SetRaw(key, -1, []byte("{\"id\":12}"))
	// bucket.SetRaw("test", -1, body)
}

func Connect() (bucket *couchbase.Bucket) {
	bucket, err := couchbase.GetBucket("http://localhost:8091/", "default", "wowitems")
	if err != nil {
		fmt.Printf("COUCHBASE FAILURE: %s\n", err)
		os.Exit(1)
	}
	return bucket
}
