build:
	go build -o bin/lb cmd/lb/main.go
	go build -o bin/be cmd/be/main.go

clean:
	rm -rf bin/