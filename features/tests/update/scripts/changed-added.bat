@ECHO OFF

@RMDIR /s /q f:\Dropbox\swap\golang\vonblog\features\tests\update\changed\media
@RMDIR /s /q f:\Dropbox\swap\golang\vonblog\features\tests\update\changed\posts
@DEL f:\Dropbox\swap\golang\vonblog\features\tests\update\changed\posts\rss.xml
@MKDIR f:\Dropbox\swap\golang\vonblog\features\tests\update\changed\media
@MKDIR f:\Dropbox\swap\golang\vonblog\features\tests\update\changed\posts
@RMDIR /s /q f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\media
@RMDIR /s /q f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts
@MKDIR f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\media
@MKDIR f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts

COPY f:\Dropbox\swap\golang\vonblog\features\tests\update\LogVisualiser.png f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\media\image.jpg
echo "<html><body>here</body></html>" > f:\Dropbox\swap\golang\vonblog\features\tests\update\changed\posts\post.html
COPY f:\Dropbox\swap\golang\vonblog\features\tests\update\scripts\changed-added.xml f:\Dropbox\swap\golang\vonblog\features\tests\update\changed\all-rss.xml
COPY f:\Dropbox\swap\golang\vonblog\features\tests\update\scripts\changed-added.xml f:\Dropbox\swap\golang\vonblog\features\tests\update\changed\rss.xml
echo --- > f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts\file.md
echo Title: "Title" >> f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts\file.md
echo Tags: [well,then] >> f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts\file.md
echo Created: "2022-04-05T22:48:15+1000" >> f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts\file.md
echo Updated: "2022-04-05T22:48:20+1000" >> f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts\file.md
echo Type: article >> f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts\file.md
echo Synopsis: "Synopsis" >> f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts\file.md
echo Slug: "file.html" >> f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts\file.md
echo --- >> f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts\file.md
echo Hi there >> f:\Dropbox\swap\golang\vonblog\features\tests\update\changed-repo\posts\file.md

echo A	media\image.jpg
echo M	posts\file.md

