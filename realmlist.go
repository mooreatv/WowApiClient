// Copyright 2019 MooreaTv moorea@ymail.com
// All Rights Reserved

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/log"
)

var (
	regionFlag = flag.String("region", "us", "Region to query (us,eu,kr,tw,cn)")
	localeFlag = flag.String("locale", "en_US", "Locale to use (en_US, es_ES, fr_FR,...)")
)

func getToken() string {
	cid := os.Getenv("OAUTH_CID")
	if len(cid) == 0 {
		log.Fatalf("Please specify client id as OAUTH_CID env var")
	}
	csec := os.Getenv("OAUTH_SEC")
	if len(csec) == 0 {
		log.Fatalf("Please specify client secret as OAUTH_SEC env var")
	}
	// curl -u {client_id}:{client_secret} -d grant_type=client_credentials https://us.battle.net/oauth/token
	tokenURL := "https://us.battle.net/oauth/token" // always US even for EU ?
	log.Infof("Getting token from %s...", tokenURL)

	o := fhttp.HTTPOptions{
		URL:               tokenURL,
		DisableFastClient: true,
		FollowRedirects:   true,
		UserCredentials:   cid + ":" + csec,
		ContentType:       "application/x-www-form-urlencoded",
		Payload:           []byte("grant_type=client_credentials"),
	}
	res, data := fhttp.Fetch(&o)
	log.LogVf("code = %d, data=%s", res, fhttp.DebugSummary(data, 512))

	var kv map[string]interface{}

	if err := json.Unmarshal(data, &kv); err != nil {
		log.Fatalf("Unable to parse json result %s: %v", fhttp.DebugSummary(data, 80), err)
	}
	token := kv["access_token"]
	log.Infof("Found token to be %v (in %v)", token, kv)
	return token.(string)
}

func main() {
	flag.Parse()
	token := getToken()
	url := fmt.Sprintf("https://%s.api.blizzard.com/wow/realm/status?locale=%s&access_token=%s", *regionFlag, *localeFlag, token)
	log.Infof("Using url %s", url)
	o := fhttp.HTTPOptions{
		URL:               url,
		DisableFastClient: true,
		FollowRedirects:   true,
	}
	res, data := fhttp.Fetch(&o)
	log.LogVf("code = %d, data=%s", res, fhttp.DebugSummary(data, 512))
	outFile := os.Stdout
	var jsonIndented bytes.Buffer
	if err := json.Indent(&jsonIndented, data, "", "  "); err != nil {
		log.Fatalf("Unable to indent json result %s: %v", fhttp.DebugSummary(data, 80), err)
	}
	jsonIndented.WriteTo(outFile)
}
