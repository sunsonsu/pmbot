go test || echo "Test failed"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go types.go
zip function.zip bootstrap  