test: ./mocks/mock_client.go ./mocks/mock_decipher.go
	go get github.com/golang/mock/gomock
	go get github.com/golang/mock/mockgen
	go test

./mocks/mock_client.go:
	mockgen --source client.go --destination mocks/mock_client.go

./mocks/mock_decipher.go:
	mockgen --source decipher.go --destination mocks/mock_decipher.go

build:
	go build -o downloader .
