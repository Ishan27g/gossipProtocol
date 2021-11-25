clean:
	rm gossiper main
build:
	go build -o gossiper gossiper.go
	go build -o main main.go
run:
	./main
runClean:
	make clean
	make RunBuild


# Build and run
RunBuild:
	make build
	make run
	
# Vendor, Clean, Build and Run
RunVendorClean:
	go mod tidy
	go mod vendor
	make runClean
