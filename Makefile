run:
	go run cmd/main.go

build:
	go build -o loyalty-points-system-api cmd/main.go

test:
	go test ./...

clean:
	rm -f loyalty-points-system-api
