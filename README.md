# restapp

[![github](https://img.shields.io/github/stars/Mopsgamer/restapp.svg?style=flat)](https://github.com/Mopsgamer/restapp)
[![github issues](https://img.shields.io/github/issues/Mopsgamer/restapp.svg?style=flat)](https://github.com/Mopsgamer/restapp/issues)

Definitely a cool chat application.

## Building from source

Requirements:

- MySQL@^8.0
- Go@^1.23
- Deno@^2.0

> [!TIP] First setup:
>
> 1. Run `deno task init:build`.
> 2. Change the `.env`.
> 3. Run `deno task serve`.

### Preparing

Creating/Updating project files:

```bash
# init sql db tables and the .env file (no overrides)
deno task init
# and build the ./web (ts, css)
deno task init:build
```

> [!NOTE]
> Fill the `.env` this file manually.

### Changing the code base

```bash
# go (server): use after *:build or *:watch
deno task serve

# web (client): use after init
deno task serve:build
deno task serve:watch
```

Use `build` or `watch` when:

- CSS or TS code is changed: `./web/src`.
- Changed any html template and potentially used new tailwind classnames.
  Otherwise, it may partially not to work.
- Deno dependencies (deno.json) are updated.

Read more about contributing [here](./CONTRIBUTING.md).
