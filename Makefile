NAME=downloaderd
MINOR_VERSION=$(shell git log -1 --format=%ad --date=short | sed 's/-//g')

VERSION=0.0.$(MINOR_VERSION)


INSTALLDIR="tmp-installdir"

all: build

build: clean
	go build

clean:
	go clean
	rm -rf *.deb
	rm -rf $(INSTALLDIR)

dummyinstall: build
	mkdir $(INSTALLDIR)
	mkdir -p $(INSTALLDIR)/usr/bin
	mkdir -p $(INSTALLDIR)/etc/downloaderd
	cp downloaderd $(INSTALLDIR)/usr/bin

package: dummyinstall
	fpm -s dir -t deb -n $(NAME) -v $(VERSION) -C $(INSTALLDIR) \
	-p $(NAME)-VERSION_ARCH.deb \
	usr/bin etc/downloaderd
