#!/bin/bash

# used to exit on first error (any non-zero exit code)
set -e

rm -rf /tmp/swisstronik-build

mkdir -p /tmp/swisstronik-build/deb/"$DEB_BIN_DIR"
cp -f ./swisstronikd /tmp/swisstronik-build/deb/"$DEB_BIN_DIR"/swisstronikd
chmod +x /tmp/swisstronik-build/deb/"$DEB_BIN_DIR"/swisstronikd

mkdir -p /tmp/swisstronik-build/deb/"$DEB_LIB_DIR"

cp -f /usr/lib/.swisstronik-enclave/* /tmp/swisstronik-build/deb/"$DEB_LIB_DIR"/
chmod +x /tmp/swisstronik-build/deb/"$DEB_LIB_DIR"/*.so

mkdir -p /tmp/swisstronik-build/deb/DEBIAN
cp deb/control /tmp/swisstronik-build/deb/DEBIAN/control
printf "Version: " >> /tmp/swisstronik-build/deb/DEBIAN/control
printf "$VERSION" >> /tmp/swisstronik-build/deb/DEBIAN/control
echo "" >> /tmp/swisstronik-build/deb/DEBIAN/control
cp deb/postinst /tmp/swisstronik-build/deb/DEBIAN/postinst
chmod 755 /tmp/swisstronik-build/deb/DEBIAN/postinst
cp deb/postrm /tmp/swisstronik-build/deb/DEBIAN/postrm
chmod 755 /tmp/swisstronik-build/deb/DEBIAN/postrm
cp deb/triggers /tmp/swisstronik-build/deb/DEBIAN/triggers
chmod 755 /tmp/swisstronik-build/deb/DEBIAN/triggers
dpkg-deb --build /tmp/swisstronik-build/deb/ .
rm -rf /tmp/swisstronik-build

cp ./swisstronik_"$VERSION"_amd64.deb /build/

