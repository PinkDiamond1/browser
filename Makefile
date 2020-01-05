SERVER_USER=zipper
SERVER_IP=192.168.1.12
SERVER_PATH=~/upload/testnet/data

CONFIG_FILE=config.yaml
MYSQL_FILE=mysql.sql
START_FILE=start.sh
STOP_FILE=stop.sh

LINUX=CGO_ENABLED=0 GOOS=linux GOARCH=amd64


## linux platform
BROWSER_LINUX=browser


linux: build/$(BROWSER_LINUX)

build:
	@mkdir -p $@

build/$(BROWSER_LINUX):
	@$(LINUX) go build -o $@ main.go
	@echo "build $@ for linux done"

upload:
	scp build/$(BROWSER_LINUX) $(SERVER_USER)@$(SERVER_IP):$(SERVER_PATH)

upload/all:
	scp build/$(BROWSER_LINUX) resetDB $(CONFIG_FILE) $(MYSQL_FILE) $(START_FILE) $(STOP_FILE) $(SERVER_USER)@$(SERVER_IP):$(SERVER_PATH)

clean:
	@echo "Cleaning binaries "
	rm -rf  build/$(BROWSER_LINUX)
