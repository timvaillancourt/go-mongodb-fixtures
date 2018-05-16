GOCACHE?=

TEST_ENABLE_DB_TESTS?=
TEST_DB_VERSION?=latest
TEST_PSMDB_PORT?=65217
TEST_MONGODB_PORT?=65218

all: get-fixtures

get-fixtures: fixtures.go cmd/get-fixtures.go
	GOCACHE=$(GOCACHE) go build -ldflags="-w -s" cmd/get-fixtures.go

$(GOPATH)/bin/gocoverutil:
	go get -u github.com/AlekSi/gocoverutil 

test-prepare: docker-compose.yml
	TEST_DB_VERSION=$(TEST_DB_VERSION) \
	TEST_PSMDB_PORT=$(TEST_PSMDB_PORT) \
	TEST_MONGODB_PORT=$(TEST_MONGODB_PORT) \
	docker-compose up -d

test: $(GOPATH)/bin/gocoverutil
	GOCACHE=$(GOCACHE) \
	TEST_ENABLE_DB_TESTS=$(TEST_ENABLE_DB_TESTS) \
	TEST_DB_VERSION=$(TEST_DB_VERSION) \
	TEST_PSMDB_PORT=$(TEST_PSMDB_PORT) \
	TEST_MONGODB_PORT=$(TEST_MONGODB_PORT) \
	gocoverutil test -v ./...

test-clean:
	docker-compose down

clean:
	rm -f cover.out get-fixtures 2>/dev/null || true
