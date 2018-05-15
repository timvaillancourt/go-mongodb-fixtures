all: get-fixtures

get-fixtures: fixtures.go cmd/get-fixtures.go
	go build -ldflags="-w -s" cmd/get-fixtures.go

clean:
	rm -f get-fixtures 2>/dev/null || true
