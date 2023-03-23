#!/bin/bash

for f in ../templates/*.twig; do
    vimdiff $f <(ssh relapse@vonexplaino.com "cat ~/cgi-bin/templates/$f")
done

for f in ../templates/h/*.twig; do
    vimdiff $f <(ssh relapse@vonexplaino.com "cat ~/cgi-bin/templates/$f")
done

for f in ../templates/share/*.twig; do
    vimdiff $f <(ssh relapse@vonexplaino.com "cat ~/cgi-bin/templates/$f")
done

for f in ../templates/webpost/*.twig; do
    vimdiff $f <(ssh relapse@vonexplaino.com "cat ~/cgi-bin/templates/$f")
done