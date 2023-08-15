#!/bin/bash

# Exit on error
set -e

# Variables
PKG_NAME="docker-shell"
PKG_VERSION="1.0.0" # This can be dynamic, for example, taken from git tags.
SOURCE_DIR="${PKG_NAME}-${PKG_VERSION}"
TARBALL="${SOURCE_DIR}.tar.gz"
MAINTAINER="Your Name <your-email@example.com>" # Adjust this.

# Prepare source tarball from the current Git repo
git archive --format=tar.gz --prefix="${SOURCE_DIR}/" -o "${TARBALL}" HEAD

# Extract the tarball
mkdir "${SOURCE_DIR}"
tar -xf "${TARBALL}" -C "${SOURCE_DIR}"

# Use dh_make to set up Debian package structure
cd "${SOURCE_DIR}"
dh_make --native -s -e "${MAINTAINER}" -f ../"${TARBALL}"

# Clean up unnecessary example files
rm -rf debian/*.ex debian/*.EX debian/README.*

# Build the package with debuild
debuild -us -uc

# Cleanup
cd ..
rm -rf "${SOURCE_DIR}"
rm "${TARBALL}"

echo "Debian package creation complete!"
