test: ./lib/mocks/mock_client.go ./lib/mocks/mock_decipher.go
	cd lib && ¬go test

./lib/mocks/mock_client.go:
	mockgen --source lib/client.go --destination lib/mocks/mock_client.go

./lib/mocks/mock_decipher.go:
	mockgen --source lib/decipher.go --destination lib/mocks/mock_decipher.go

build:
	go build -o downloader .
