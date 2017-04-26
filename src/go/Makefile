all:
	ln -s `pwd`/appscope.net $$GOPATH/src/appscope.net
	(cd appscope.net; make)
	mkdir -p ../build
	CGO_LDFLAGS=-lnuma go build -ldflags "-X main.version=${VERSION}" -o ../build/capture-server appscope.net/capture/daemon
	
