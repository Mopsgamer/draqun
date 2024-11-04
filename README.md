# restapp

[![github](https://img.shields.io/github/stars/Mopsgamer/restapp.svg?style=flat)](https://github.com/Mopsgamer/restapp)
[![github issues](https://img.shields.io/github/issues/Mopsgamer/restapp.svg?style=flat)](https://github.com/Mopsgamer/restapp/issues)

Definitely a cool chat application.

## Building from source

Requirements:
- MySql@>=2
- go@>=1.23

Optional:
- Deno@>=2 (optional, for web rebuilds)

### Preparing

Creating the `.env` file:
```bash
go run . --make-env
# then fill it by in your editor
```

Bundling the web (optional, use only when updating deno dependencies):
```bash
deno task build
```

### Starting the server

Running the server
```bash
go run .
# or
deno task serve
```
