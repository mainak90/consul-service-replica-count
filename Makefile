.PHONY: build release

build:
		rm -rf build/
		gox -os="linux darwin" -arch="amd64" -output="build/service-replica-count"

release:
		rm -rf release/
		gox -os="linux darwin" -arch="amd64" -output="release/service-replica-count_{{.OS}}_{{.Arch}}"
		gh-release checksums sha256