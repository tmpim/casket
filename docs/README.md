# casket-docs

This directory contains the documentation for [Casket](https://github.com/tmpim/casket), a fork of [mholt's Caddy web
server](https://github.com/caddyserver/caddy) v1. This documentation is forked from the Caddy v1 docs as of June 2020,
converted from HTML to Markdown in [this
commit](https://github.com/tmpim/casket-docs/commit/2ac4b866cad687f52a94d262ee2687a5b4e60acd), and then updated to
reflect the changes in Casket.

For more information on what Casket is, see the [Casket repository](https://github.com/tmpim/casket), or the [Casket
documentation](https://docs.casketserver.com/).

## Contributing

If you find a typo or other error in the documentation, please open an issue or pull request. The documentation is
written in [Markdown](https://commonmark.org/help/) and is located in the `src` directory.

## Building

The documentation is built with [VitePress](https://vitepress.dev/), and is compatible with VitePress's [markdown
extensions](https://vitepress.dev/guide/markdown). To build the documentation and preview it in your browser, you will
need to install [Node.js](https://nodejs.org/en/). Then, run the following commands:

```shell
# Set up the repository
git clone https://github.com/tmpim/casket.git
cd casket/docs

# Install the dependencies
pnpm install

# Preview the documentation
pnpm dev
```
