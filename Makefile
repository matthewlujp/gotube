test: ./lib/mocks/mock_client.go ./lib/mocks/mock_decipher.go
	go get github.com/golang/mock/gomock
	go get github.com/golang/mock/mockgen
	cd lib && go test

./lib/mocks/mock_client.go:
	mockgen --source lib/client.go --destination lib/mocks/mock_client.go

./lib/mocks/mock_decipher.go:
	mockgen --source lib/decipher.go --destination lib/mocks/mock_decipher.go

build:
	go build -o downloader .
