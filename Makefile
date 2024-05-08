.PHONY: build release

binary = hload
main = main.go
os = $(shell uname -s | tr '[:upper:]' '[:lower:]')
arch = $(shell uname -m)
git_commit = $(shell git rev-parse --short HEAD)
major = 1
minor = 2
patch = 0
go_version = $(shell go version | awk '{print $$3}')
build_date = $(shell date +%FT%T%z)

all: build amd64 linux

build:
	@echo "Building native..."
	@CGO_ENABLED=1 go build -o bin/$(binary)-$(os)-$(arch)-v$(major).$(minor).$(patch) -ldflags "-X 'github.com/tmwalaszek/hload/cmd/version.GitCommit=$(git_commit)' \
		-X 'github.com/tmwalaszek/hload/cmd/version.Major=$(major)' \
		-X 'github.com/tmwalaszek/hload/cmd/version.Minor=$(minor)' \
		-X 'github.com/tmwalaszek/hload/cmd/version.Patch=$(patch)' \
		-X 'github.com/tmwalaszek/hload/cmd/version.GoVersion=$(go_version)' \
		-X 'github.com/tmwalaszek/hload/cmd/version.BuildDate=$(build_date)'" $(main)

amd64:
	@echo "Building amd64..."
	@GOARCH=amd64 CGO_ENABLED=1 go build -o bin/$(binary)-$(os)-amd64-v$(major).$(minor).$(patch) -ldflags "-X 'github.com/tmwalaszek/hload/cmd/version.GitCommit=$(git_commit)' \
        -X 'github.com/tmwalaszek/hload/cmd/version.Major=$(major)' \
        -X 'github.com/tmwalaszek/hload/cmd/version.Minor=$(minor)' \
        -X 'github.com/tmwalaszek/hload/cmd/version.Patch=$(patch)' \
        -X 'github.com/tmwalaszek/hload/cmd/version.GoVersion=$(go_version)' \
        -X 'github.com/tmwalaszek/hload/cmd/version.BuildDate=$(build_date)'" $(main)

linux:
	@echo "Building for linux..."
	@docker build -t hload-build --platform=linux/amd64 .
	@docker create --name hload-linux --platform=linux/amd64 hload-build
	@docker cp hload-linux:/app/bin/$(binary)-linux-amd64-v$(major).$(minor).$(patch) bin/$(binary)-linux-amd64-v$(major).$(minor).$(patch)
	@docker rm -f hload-linux
	@docker rmi hload-build

