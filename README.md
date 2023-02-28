# blogsync

First draft for a simple long-form text note ([NIP-23][nip23]) sync to article folder.

Run like this:

`go run main.go example_conf.json`

The article folder can then e.g. be served via [Caddy server][caddy].
See [caddy_example][example] for an example.

Sync will stop after EOS, so you have to run periodically, e.g. via cron.

## TODOs

- Wrapping this into a [Caddy server][caddy] plugin would be nice.
- See code for further TODOs.

## Feedback

npub1hz404lnj7lxsdt5vxdlex9rlvhlz6dxqqed4y6tpywvzgwx0qmlqfpl6sm

[caddy]: https://caddyserver.com/
[example]: caddy_example
[nip23]: https://github.com/nostr-protocol/nips/blob/master/23.md
