# WowApiClient
WoW API (https://develop.battle.net/) client code in GoLang

## How to use:
Create a client on https://develop.battle.net/access/clients
and set `OAUTH_CID` and `OAUTH_SEC` in your environment, the program will also try to use `OAUTH_{CID|SEC}_{REGION}` if present. Use `-authregion eu` to use eu battle.net oauth client instead of us.

Then `go run realmlist.go -status` for US server status list and `go run realmlist.go -region eu` for EU region list

By default it pretty prints the whole Json structure returned (which is more complete when using the `-status` realm status api option) but you can use `-lua` to just get the realm id to names as LUA code suitable for WoW Addons (use https://github.com/mooreatv/MoLib if you just want to use that data)
```Shell
go run realmlist.go -lua -region eu > Realms_eu.lua
go run realmlist.go -lua -region us > Realms_us.lua
```

Add `-loglevel verbose` or `-loglevel debug` to see more details during the execution.
