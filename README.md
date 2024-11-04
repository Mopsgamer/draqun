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

Creating project files:
```bash
go run . -- --init
# or
deno task init
```

After this command you will get the `.env` file in your project's root.
Fill this file manually.

Force option overrides your existing files:
```bash
go run . -- --init --force
# or
deno task init --force
```

### Starting the server

Running the server
```bash
go run .
# or
deno task serve
```

### Changing the code base

Bundling the web statics and assets:
```bash
deno task build
```

The command above should be used when:

- Repository has been cloned.
- Deno dependencies (js libraries) has been updated.
- Changed any html template and potentially used new tailwind classnames, Otherwise it may not work partially.
- CSS or JS code has been changed: `./web/src`.
