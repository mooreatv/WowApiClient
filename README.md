# WowApiClient
WoW API (https://develop.battle.net/) client code in GoLang

## How to use:
Create a client on https://develop.battle.net/access/clients
and set `OAUTH_CID` and `OAUTH_SEC` in your environment

Then `go run realmlist.go` for US server list and `go run realmlist.go -region eu` for EU region list
