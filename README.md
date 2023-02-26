# blogsync

First draft for a simple long-form text note ([NIP-23][nip23]) sync to article folder.

Provide author PubKeys as command line args, e.g.:

`go run main.go b8aafafe72f7cd06ae8c337f93147f65fe2d34c0065b52696123982438cf06fe`

The article folder can then e.g. be served via [Caddy server][caddy].

Sync will stop after EOS, so you have to run periodically, e.g. via cron.

## TODOs

- Wrapping this into a [Caddy server][caddy] plugin would be nice.
- See code for further TODOs.

## Feedback

npub1hz404lnj7lxsdt5vxdlex9rlvhlz6dxqqed4y6tpywvzgwx0qmlqfpl6sm

[caddy]: https://caddyserver.com/
[nip23]: https://github.com/nostr-protocol/nips/blob/master/23.md
