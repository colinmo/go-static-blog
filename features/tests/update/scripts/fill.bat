@ECHO OFF

RMDIR /s /q c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\media
RMDIR /s /q c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts
MKDIR c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\media
MKDIR c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts

COPY c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\LogVisualiser.png c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\media\
echo --- > c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts\file.md
echo Title: "Title" >> c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts\file.md
echo Tags: [well,then] >> c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts\file.md
echo Created: "2022-04-05T22:48:15+1000" >> c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts\file.md
echo Updated: "2022-04-05T22:48:20+1000" >> c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts\file.md
echo Type: article >> c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts\file.md
echo Synopsis: "Synopsis" >> c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts\file.md
echo FeatureImage: /blog/media/FeatureImage >> c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts\file.md
echo --- >> c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts\file.md
echo Some content, I guess >> c:\users\relap\dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\posts\file.md

echo Hi there
echo A	media\image.jpg
echo M	posts\file.md
echo wow

