# WowApiClient
WoW API (https://develop.battle.net/) client code in GoLang

## How to use:
Create a client on https://develop.battle.net/access/clients
and set `OAUTH_CID` and `OAUTH_SEC` in your environment

Then `go run realmlist.go` for US server list and `go run realmlist.go -region eu` for EU region list

By default it pretty prints the whole Json structure returned but you can use `-names` to just get the realm names in a list ready to be inserted into a LUA table
```Shell
go run realmlist.go -names -region eu > Realms_eu.lua
go run realmlist.go -names -region us > Realms_us.lua
```
The resulting files and utility LUA functions for Wow addons are available in as part of https://github.com/mooreatv/MoLib

Add `-loglevel verbose` or `-loglevel debug` to see more details during the execution.
