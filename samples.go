package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var port int = 8093
var templatePath string
var contentMux = http.NewServeMux()
var bucket *couchbase.Bucket
var casb *couchbase.Bucket

func main() {
	fmt.Println("Starting server")

	api.Domain = "192.168.1.159"
	bucket = Connect("wowitems")
	casb = Connect("wowitems-castest")

	Start("./")
}

type Results struct {
	Query         string
	SearchResults *core.SearchResult
	WoWItems      []*WoWItemDoc
	CasQueryType  int
	IsHome        bool
	IsSearch      bool
	IsCas         bool
	Duration      time.Duration
	CacheDoc      *WoWItem
}

func (r Results) HasCacheDoc() bool {
	return r.CacheDoc != nil && r.CacheDoc.Id > 0
}

type WoWItemDoc struct {
	Doc  WoWItem `json:"doc"`
	Meta Meta    `json:"meta"`
}

type Meta struct {
	Id string `json:"id"`
}

type WoWItem struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	RawJson     string
	CasValue    uint64
	CacheHit    bool
}

func Start(root string) {
	templatePath = fmt.Sprintf("%ssamples/index.html", root)

	contentMux.Handle("/", http.FileServer(http.Dir(fmt.Sprintf("%ssamples/", root))))

	fmt.Printf("Listing on localhost:%d\n", port)
	http.HandleFunc("/", handleRequest)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if err != nil {
		fmt.Printf("ListenAndServe Error :%s\n", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Handling request %s\n", r.URL)
	urlString := r.URL.String()
	switch {
	case strings.HasPrefix(urlString, "/cas"):
		handleCasRequest(w, urlString)
	default:
		handleResultsRequest(w, urlString)
	}
}

func handleResultsRequest(w http.ResponseWriter, urlString string) {
	t0 := time.Now()
	var q string = ""

	if urlString != "" {
		u, _ := url.Parse(urlString)
		q = u.Query().Get("q")
	}
	t := template.New("index.html")
	t, err := t.ParseFiles(templatePath)
	if err != nil {
		fmt.Fprintf(w, "ERROR %s", err)
		return
	}
	r := Results{Query: q}

	if q != "" {
		searchJson := `{
	        "query" : {
	            "bool" : { "must" : [
	            	{
	            		"query_string": {
	            			"default_field" : "_all",
	            			"query" : "` + q + `"
	            		}
	            	}
	             ]
	        	}
	        }
	    }`
		out, _ := core.SearchRequest(true, "wowitems", "couchbaseDocument", searchJson, "", 0)
		r.SearchResults = &out
		wowItems := make([]*WoWItemDoc, 0)

		for _, hit := range out.Hits.Hits {
			dec := json.NewDecoder(bytes.NewReader(hit.Source))
			// fmt.Printf("ORIG=%s\n", string(hit.Source))
			var itm WoWItemDoc
			err = dec.Decode(&itm)
			if err != nil {
				fmt.Printf("ERROR %s", err)
			}
			fmt.Printf("%s\n", itm)
			itm.Doc.RawJson = string(hit.Source)
			wowItems = append(wowItems, &itm)
		}

		r.WoWItems = wowItems
		r.IsSearch = true
		// fmt.Printf("Found %s\n", len(out.Hits.Hits))
	} else {
		r.IsHome = true
	}
	r.Duration = time.Now().Sub(t0)
	err = t.Execute(w, r)
	if err != nil {
		fmt.Fprintf(w, "ERROR %s", err)
		return
	}
}

func handleCasRequest(w http.ResponseWriter, urlString string) {
	t := template.New("index.html")
	t, err := t.ParseFiles(templatePath)
	if err != nil {
		fmt.Fprintf(w, "ERROR %s", err)
		return
	}

	var q string = ""
	var id int

	if urlString != "" {
		u, _ := url.Parse(urlString)
		q = u.Query().Get("qt")
		id, _ = strconv.Atoi(u.Query().Get("id"))
	}

	if id == 0 {
		id = 1
	}

	ty, _ := strconv.Atoi(q)
	r := Results{CasQueryType: ty, IsCas: true}

	if r.CasQueryType == 2 {
		CreateRandDoc(id)
	}

	doc := GetLocalCache(fmt.Sprintf("item_%d", id))
	r.CacheDoc = doc

	err = t.Execute(w, r)
	if err != nil {
		fmt.Fprintf(w, "ERROR %s", err)
		return
	}
}

var r *rand.Rand = rand.New(rand.NewSource(99))

func CreateRandDoc(id int) {
	fmt.Printf("Creating item_%d\n", id)
	key := fmt.Sprintf("item_%d", id)
	name := fmt.Sprintf("Random Name %d", r.Int())
	desc := fmt.Sprintf("Random Description %d", r.Int63())
	wi := &WoWItem{Id: id, Name: name, Description: desc}
	casb.Set(key, -1, wi)
}

type CacheItem struct {
	LastCas uint64
	Doc     *WoWItem
}

var cache map[string]*CacheItem = make(map[string]*CacheItem)

func GetLocalCache(key string) *WoWItem {
	o, e := casb.Observe(key) // we're going to do this no matter what
	if e != nil {
		fmt.Printf("--> cache miss / Observe error on key %s err is %s", key, e)
	}
	if ci, hit := cache[key]; hit && ci.LastCas == o.Cas {
		fmt.Printf("CACHE HIT! %s\n", key)
		ci.Doc.CacheHit = true
		return ci.Doc
	}

	fmt.Printf("--> cache miss %s\n", key)

	var wi WoWItem
	e = casb.Get(key, &wi)
	if e != nil {
		fmt.Printf("ERROR %s\n", e)
		return nil
	}
	wi.CasValue = o.Cas
	cache[key] = &CacheItem{LastCas: o.Cas, Doc: &wi}
	return &wi
}

func Connect(n string) (bucket *couchbase.Bucket) {
	bucket, err := couchbase.GetBucket("http://localhost:8091/", "default", n)
	if err != nil {
		fmt.Printf("COUCHBASE FAILURE: %s\n", err)
		os.Exit(1)
	}
	return bucket
}
