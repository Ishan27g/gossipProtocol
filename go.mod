module github.com/Ishan27gOrg/gossipProtocol

go 1.17

replace (
	github.com/Ishan27gOrg/gossipProtocol => ../gossipProtocol
	github.com/Ishan27gOrg/raftProtocol => ../raftProtocol
	github.com/Ishan27gOrg/registry/package => ../registry/package
)

require (
	github.com/Ishan27g/go-utils/mLogger v0.0.0-20211124154642-ddf1831bec07
	github.com/Ishan27g/go-utils/pEnv v0.0.0-20211124154642-ddf1831bec07
	github.com/Ishan27gOrg/vClock v0.0.0-00010101000000-000000000000
	github.com/emirpasic/gods v1.12.0
	github.com/hashicorp/go-hclog v1.0.0
	github.com/stretchr/testify v1.7.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20191008105621-543471e840be // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/Ishan27gOrg/vClock => ../vClock
