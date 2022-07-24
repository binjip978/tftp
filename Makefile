.PHONY: all
all: test client server

.PHONY: client
client:
	cd cmd/tftp-client && go build -o ../../tftp-client

.PHONY: server
server:
	cd cmd/tftp-server && go build -o ../../tftp-server

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -f tftp-server tftp-client
