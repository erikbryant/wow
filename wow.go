package main

import (
	"./web"
	"flag"
	"fmt"
)

var (
	clientID     = flag.String("clientID", "", "WoW API client ID")
	clientSecret = flag.String("clientSecret", "", "WoW API client secret")
)

// https://develop.battle.net/documentation

func accessToken(id, secret string) string {
	url := "https://us.battle.net/oauth/token?client_id=" + id + "&client_secret=" + secret + "&grant_type=client_credentials"
	response := web.WebRequestJSON(url)
	return response["access_token"].(string)
}

func auctionHouse(realm, accessToken string) []interface{} {
	// Look up the URL for the latest auction house data.
	url := "https://us.api.blizzard.com/wow/auction/data/" + realm + "?locale=en_US&access_token=" + accessToken
	response := web.WebRequestJSON(url)
	data := response["files"].([]interface{})[0].(map[string]interface{})
	auctionURL := data["url"].(string)

	// Retrieve the latest auction house data.
	response = web.WebRequestJSON(auctionURL)
	// realms := response["realms"].([]interface{})
	auctions := response["auctions"].([]interface{})
	return auctions
}

func item(id, accessToken string) map[string]interface{} {
	url := "https://us.api.blizzard.com/wow/item/" + id + "?locale=en_US&access_token=" + accessToken
	return web.WebRequestJSON(url)
}

func main() {
	flag.Parse()

	realm := "icecrown"
	token := accessToken(*clientID, *clientSecret)

	auctions := auctionHouse(realm, token)
	fmt.Println(auctions[0])
	fmt.Println()

	first := auctions[0].(map[string]interface{})
	i := item(web.ToString(first["item"]), token)
	fmt.Println(i)
	fmt.Println()
}
