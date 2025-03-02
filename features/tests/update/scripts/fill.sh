#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
rm -fr ${SCRIPT_DIR}/../fullregenrep/media
rm -fr ${SCRIPT_DIR}/../fullregenrep/posts
mkdir ${SCRIPT_DIR}/../fullregenrep/media
mkdir ${SCRIPT_DIR}/../fullregenrep/posts

cp "${SCRIPT_DIR}/../LogVisualiser.png" "${SCRIPT_DIR}/../fullregenrep/media/"
FILENAME="${SCRIPT_DIR}/../fullregenrep/posts/file.md"
echo --- > $FILENAME
echo --- > c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts\file.md
echo Title: "Title" >> $FILENAME
echo Tags: [well,then] >> $FILENAME
echo Created: "2022-04-05T22:48:15+1000" >> $FILENAME
echo Updated: "2022-04-05T22:48:20+1000" >> $FILENAME
echo Type: article >> $FILENAME
echo Synopsis: "Synopsis" >> $FILENAME
echo FeatureImage: /blog/media/FeatureImage >> $FILENAME
echo --- >> $FILENAME
echo Some content, I guess >> $FILENAME

echo Hi there
echo A	media\image.jpg
echo M	posts\file.md
echo wow

