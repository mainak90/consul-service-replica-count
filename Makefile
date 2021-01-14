.PHONY: build release

build:
		rm -rf build/
		go build -o build/service-replica-count main.go
		cp -f build/service-replica-count /usr/local/bin/service-replica-count

release:
		rm -rf release/
		go build -os="linux-darwin" -arch="amd64" -o="release/service-replica-count_{{.OS}}_{{.Arch}}" main.go
		gh-release checksums sha256