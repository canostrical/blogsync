# Caddy Setup Example for blogsync

Most of the magic is happening in [Caddyfile](Caddyfile) and [index.html](homepage/index.html).

Running blogsync with [example config](../example_conf.json) should create markdown files in homepage/markdown.

Running Caddy (`caddy run`) within the homepage folder should then serve rendered versions.

E.g. if there is a file `markdown/some-d-tag.md` it should be available under http://localhost:2019/articles/some-d-tag.