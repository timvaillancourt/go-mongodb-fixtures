all: get-fixtures

get-fixtures: fixtures.go cmd/get-fixtures.go
	go build -ldflags="-w -s" cmd/get-fixtures.go

$(GOPATH)/bin/gocoverutil:
	go get -u github.com/AlekSi/gocoverutil 

test: $(GOPATH)/bin/gocoverutil
	gocoverutil test -v ./...

clean:
	rm -f cover.out get-fixtures 2>/dev/null || true
