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

Creating the `.env` file:
```bash
go run . -- --make-env
# or
deno task serve -- --make-env

# then fill it by in your editor
```

Bundling the web statics and assets:
```bash
deno task build
```

The command above should be used when:

- Repository has been cloned.
- Deno dependencies (js libraries) has been updated.
- Changed any html template and potentially used new tailwind classnames, Otherwise it may not work partially.
- CSS or JS code has been changed: `./web/src`.

### Starting the server

Running the server
```bash
go run .
# or
deno task serve
```
