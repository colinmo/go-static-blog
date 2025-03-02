#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
rm -fr ${SCRIPT_DIR}/../changed/media
rm -fr ${SCRIPT_DIR}/../changed/posts
rm ${SCRIPT_DIR}/./changed/posts/rss.xml
mkdir ${SCRIPT_DIR}/../changed/media
mkdir ${SCRIPT_DIR}/../changed/posts

rm -fr ${SCRIPT_DIR}/../changed-repo/media
rm -fr ${SCRIPT_DIR}/../changed-repo/posts
mkdir ${SCRIPT_DIR}/../changed-repo/posts
mkdir ${SCRIPT_DIR}/../changed-repo/media

cp ${SCRIPT_DIR}/../LogVisualiser.png ${SCRIPT_DIR}/../changed-repo/media/image.jpg
echo "<html><body>here</body></html>" > ${SCRIPT_DIR}/../changed/posts/post.html
cp ${SCRIPT_DIR}/changed-added.xml ${SCRIPT_DIR}/../changed/all-rss.xml
cp ${SCRIPT_DIR}/changed-added.xml ${SCRIPT_DIR}/../changed/rss.xml
MD_FILE=${SCRIPT_DIR}/../changed-repo/posts/file.md
echo --- > ${MD_FILE}
echo Title: "Title" >> ${MD_FILE}
echo Tags: [well,then] >> ${MD_FILE}
echo Created: "2022-04-05T22:48:15+1000" >> ${MD_FILE}
echo Updated: "2022-04-05T22:48:20+1000" >> ${MD_FILE}
echo Type: article >> ${MD_FILE}
echo Synopsis: "Synopsis" >> ${MD_FILE}
echo Slug: "file.html" >> ${MD_FILE}
echo --- >> ${MD_FILE}
echo Hi there >> ${MD_FILE}

echo A	media/image.jpg
echo M	posts/file.md

