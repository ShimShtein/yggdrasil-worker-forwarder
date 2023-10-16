PKGNAME := yggdrasil-worker-forwarder

ifeq ($(origin VERSION), undefined)
	VERSION := 0.0.3
endif

.PHONY: build
build:
	mkdir -p _build
	go build -o _build/yggdrasil-worker-forwarder main.go server.go

clean:
	rm -rf _build

distribution-tarball:
	go mod vendor
	tar --create \
		--gzip \
		--file /tmp/$(PKGNAME)-$(VERSION).tar.gz \
		--exclude=.git \
		--exclude=.vscode \
		--exclude=.github \
		--exclude=.gitignore \
		--exclude=.copr \
		--transform s/^\./$(PKGNAME)-$(VERSION)/ \
		. && mv /tmp/$(PKGNAME)-$(VERSION).tar.gz .
	rm -rf ./vendor

test:
	go test *.go

vet:
	go vet *.go
