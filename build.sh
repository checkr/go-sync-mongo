#
# Build all Versions of Up in the bin directory
#
GOOS=linux GOARCH=amd64 go build -o bin/linux/go-sync-mongo
GOOS=darwin GOARCH=amd64 go build -o bin/macos/go-sync-mongo
GOOS=windows GOARCH=amd64 go build -o bin/windows/go-sync-mongo.exe
