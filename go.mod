module github.com/Ishan27gOrg/gossipProtocol

go 1.17

replace (
	github.com/Ishan27gOrg/gossipProtocol => ../gossipProtocol
	github.com/Ishan27gOrg/raftProtocol => ../raftProtocol
	github.com/Ishan27gOrg/registry/package => ../registry/package
)

require (
	github.com/Ishan27g/go-utils/mLogger v0.0.0-20211230135645-5a7e5725a982
	github.com/Ishan27g/go-utils/rtimer v0.0.0-20220102105519-ebd3e3be9b12
	github.com/Ishan27gOrg/vClock v0.0.0-00010101000000-000000000000
	github.com/emirpasic/gods v1.12.0
	github.com/hashicorp/go-hclog v1.0.0
	github.com/stretchr/testify v1.7.0
)

require (
	github.com/cloudevents/sdk-go/v2 v2.6.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/iancoleman/orderedmap v0.2.0 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/sys v0.0.0-20191008105621-543471e840be // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/Ishan27gOrg/vClock => ../vClock
