#!/bin/bash

# Build both Chinese and English documentation sites

echo "Building Chinese documentation..."
mkdocs build -f mkdocs-zh.yml

echo "Building English documentation..."
mkdocs build -f mkdocs-en.yml

echo "Done!"
echo ""
echo "Chinese site: site-zh/"
echo "English site: site-en/"
echo ""
echo "To serve locally:"
echo "  mkdocs serve -f mkdocs-zh.yml  (Chinese)"
echo "  mkdocs serve -f mkdocs-en.yml  (English)"