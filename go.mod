module github.com/Ishan27gOrg/gossipProtocol

go 1.17

replace (
	github.com/Ishan27gOrg/gossipProtocol => ../gossipProtocol
	github.com/Ishan27gOrg/raftProtocol => ../raftProtocol
	github.com/Ishan27gOrg/registry/package => ../registry/package
)

require (
	github.com/Ishan27g/go-utils/mLogger v0.0.0-20220111231648-a0642517a586
	github.com/Ishan27gOrg/vClock v0.0.0-00010101000000-000000000000
	github.com/emirpasic/gods v1.12.0
	github.com/go-playground/assert v1.2.1
	github.com/hashicorp/go-hclog v1.0.0
	github.com/jedib0t/go-pretty/v6 v6.2.7
	github.com/stretchr/testify v1.7.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/iancoleman/orderedmap v0.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	golang.org/x/sys v0.0.0-20191008105621-543471e840be // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/Ishan27gOrg/vClock => ../vClock
