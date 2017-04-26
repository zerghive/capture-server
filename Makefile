GO_PACKAGE = go1.5.1.linux-amd64.tar.gz

VERSION_MAJOR = 1
VERSION_MINOR = 0
VERSION_RELEASE = rc1

captureserver: dist/$(GO_PACKAGE)
	if [[ -z `(docker images | grep "appscope/cs_base" | awk '{print $$3}')` ]]; then docker build -f Dockerfile.cs_base -t appscope/cs_base . ; fi
	docker run -i -v `pwd`/src:/build appscope/cs_base /bin/bash -c "(cd /build/go; make VERSION="$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_RELEASE).$(shell git rev-list --first-parent --count master)")"

# force rebuild cs_base
cs_base:
	docker build --no-cache -t appscope/cs_base -f Dockerfile.cs_base .
	# cleanup intermediate containers
	docker rmi -f $$(docker images | grep "^<none>" | awk '{print $$3}')

cs_test:
	docker build --no-cache -t appscope/cs_test -f Dockerfile.cs_test .

run_local:
	docker run -i -v `pwd`/src/build:/build --privileged -p 1194:1194/udp appscope/cs_test

dist/$(GO_PACKAGE):
	mkdir -p dist
	curl -o dist/$(GO_PACKAGE) https://storage.googleapis.com/golang/$(GO_PACKAGE)
