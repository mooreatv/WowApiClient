// Copyright 2019 MooreaTv moorea@ymail.com
// All Rights Reserved

// Get your OAUTH_CID and OAUTH_SEC by creating a client on
// https://develop.battle.net/access/clients

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/log"
)

var (
	regionFlag      = flag.String("region", "us", "Region to query (us,eu,kr,tw,cn)")
	authRegionFlag  = flag.String("authregion", "us", "Battle.net to use for token")
	localeFlag      = flag.String("locale", "en_US", "Locale to use (en_US, es_ES, fr_FR,...)")
	genLuaFlag      = flag.Bool("lua", false, "Generates id to name LUA table")
	realmStatusFlag = flag.Bool("status", false, "Use the realm status API instead of realm index one")
)

// returns a region specific value or base value if region is not found
// eg OAUTH_CID_EU for eu region
func getEnv(key, region string) string {
	reg := "_" + strings.ToUpper(region)
	regKey := key + reg
	val := os.Getenv(regKey)
	if len(val) > 0 {
		log.Infof("Found and using region specific credential %s", regKey)
		return val
	}
	return os.Getenv(key)
}

func getToken(region string) string {
	cid := getEnv("OAUTH_CID", region)
	if len(cid) == 0 {
		log.Fatalf("Please specify client id as OAUTH_CID env var (https://develop.battle.net/access/clients)")
	}
	csec := getEnv("OAUTH_SEC", region)
	if len(csec) == 0 {
		log.Fatalf("Please specify client secret as OAUTH_SEC env var")
	}
	tokenURL := fmt.Sprintf("https://%s.battle.net/oauth/token", region)
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

// LocaleNameMap is Name/Locale structure
type LocaleNameMap map[string]string

// Realm structure
type Realm struct {
	Name LocaleNameMap
	ID   int
	Slug string
}

func parseRealmList(data []byte) []Realm {
	type RealmList struct {
		Realms []Realm
	}
	var realmList RealmList
	if err := json.Unmarshal(data, &realmList); err != nil {
		log.Fatalf("Unable to unmarshal json result %s: %v", fhttp.DebugSummary(data, 80), err)
	}
	log.LogVf("parsed: %#v", realmList)
	return realmList.Realms
}

func getRealmStatusList(region, token string) []byte {
	url := fmt.Sprintf("https://%s.api.blizzard.com/wow/realm/status?access_token=%s", region, token)
	log.Infof("Using url %s", url)
	o := fhttp.HTTPOptions{
		URL:               url,
		DisableFastClient: true,
		FollowRedirects:   true,
	}
	res, data := fhttp.Fetch(&o)
	log.LogVf("code = %d, data=%s", res, fhttp.DebugSummary(data, 512))
	return data
}

func getRealmList(region, token string) []byte {
	url := fmt.Sprintf("https://%[1]s.api.blizzard.com/data/wow/realm/index?namespace=dynamic-%[1]s&access_token=%s", region, token)
	log.Infof("Using url %s", url)
	o := fhttp.HTTPOptions{
		URL:               url,
		DisableFastClient: true,
		FollowRedirects:   true,
	}
	res, data := fhttp.Fetch(&o)
	log.LogVf("code = %d, data=%s", res, fhttp.DebugSummary(data, 512))
	return data

}

var header = `-- Realm list generated on %s
-- by https://github.com/mooreatv/WowApiClient
-- go run realmlist.go -lua > Realms.lua
Realms = {
`
var footer = `}
-- end of generated realm list
`

func generateLua(token string) {
	// https://develop.battle.net/documentation/guides/regionality-partitions-and-localization
	regions := []string{"us", "eu", "kr", "tw"}
	outFile := os.Stdout
	fmt.Fprintf(outFile, header, time.Now().UTC().Format(time.UnixDate))
	id2data := make(map[int]Realm)
	max := 0
	for _, region := range regions {
		data := getRealmList(region, token)
		realms := parseRealmList(data)
		// For EU seems like pt_BR has the correct realm name in russian servers and us servers too
		// without having to do 1 by 1 realm calls to find the correct locale
		locale := "pt_BR" // works for US and EU including Russia (picks the russian canonical name)
		if region == "tw" {
			locale = "zh_TW"
		} else if region == "kr" {
			locale = "ko_KR"
		}
		for _, r := range realms {
			if r.ID > max {
				max = r.ID
			}
			if o, ok := id2data[r.ID]; ok {
				log.Fatalf("Duplicate entries for %d: %v vs %v", r.ID, o, r)
			}
			r.Name["canonical"] = r.Name[locale]
			r.Name["region"] = region
			id2data[r.ID] = r
		}
	}
	holes := 0
	valid := 0
	for i := 1; i <= max; i++ {
		r, ok := id2data[i]
		if !ok {
			holes = holes + 1
			continue
		}
		valid = valid + 1
		fmt.Fprintf(outFile, "  [%d] = {\"%s\", \"%s\"}, -- \"%s\", %s\n", r.ID, r.Name["canonical"], r.Name["region"], r.Name["en_US"], r.Slug)
	}
	fmt.Fprintf(outFile, footer)
	log.Infof("Last realm id is %d, %d holes, %d valid realms found", max, holes, valid)
}

func main() {
	flag.Parse()
	region := strings.ToLower(*regionFlag)
	token := getToken(*authRegionFlag)
	var data []byte
	if *genLuaFlag {
		generateLua(token)
		return
	}
	// Pretty print everything:
	if *realmStatusFlag {
		data = getRealmStatusList(region, token)
	} else {
		data = getRealmList(region, token)
	}
	var jsonIndented bytes.Buffer
	if err := json.Indent(&jsonIndented, data, "", "  "); err != nil {
		log.Fatalf("Unable to indent json result %s: %v", fhttp.DebugSummary(data, 80), err)
	}
	outFile := os.Stdout
	jsonIndented.WriteTo(outFile)
}
