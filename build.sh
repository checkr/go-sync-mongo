#
# Build all Versions of Up in the bin directory
#
GOOS=linux GOARCH=amd64 go build -o bin/go-sync-mongo-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o bin/go-sync-mongo-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o bin/go-sync-mongo-windows-amd64.exe
