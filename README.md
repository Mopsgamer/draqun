# restapp

[![github](https://img.shields.io/github/stars/Mopsgamer/restapp.svg?style=flat)](https://github.com/Mopsgamer/restapp)
[![github issues](https://img.shields.io/github/issues/Mopsgamer/restapp.svg?style=flat)](https://github.com/Mopsgamer/restapp/issues)

Definitely a cool chat application.

## Building from source

Requirements:

- MySql@>=5.0
- go@>=1.23
- Deno@>=2.0

### Preparing

Creating/Updating project files:

```bash
deno task init
```

After this command you will get the `.env` file in your project's root. Fill
this file manually.

### Changing the code base

```bash
# go (server)
deno task serve

# web (client)
deno task serve:build
deno task serve:watch
```

Use `build` or `watch` when:

- CSS or TS code has been changed: `./web/src`.
- Changed any html template and potentially used new tailwind classnames.
  Otherwise, it may not work partially.
- Deno dependencies (deno.json) has been updated.

Read more about contributing [here](./CONTRIBUTING.md).
