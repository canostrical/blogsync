# Caddy Setup Example for blogsync

Most of the magic is happening in [Caddyfile](Caddyfile) and [index.html](homepage/index.html).

Running blogsync with [example config](../example_conf.json) should create markdown files in homepage/markdown.

Running Caddy (`caddy run`) should serve rendered articles.

To compile stylesheets, install [Tailwind CSS](https://tailwindcss.com) and dependencies (`yarn`) and run `yarn run build:css`.