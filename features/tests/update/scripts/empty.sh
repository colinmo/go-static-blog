#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
rm -fr ${SCRIPT_DIR}/../fullregenrep/media
rm -fr ${SCRIPT_DIR}/../fullregenrep/posts
mkdir ${SCRIPT_DIR}/../fullregenrep/media
mkdir ${SCRIPT_DIR}/../fullregenrep/posts
echo
