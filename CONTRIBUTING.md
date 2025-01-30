# Contributing

## Changing the code base

The `watch` and `build` client code base (./web) is not tied with the server
code base (./internal). The best way is to use 2 terminals (3-rd for other
tasks):

```bash
deno task serve --watch
```

```bash
deno task watch
```

> [!NOTE]
> You can use Visual Studio Code's task commands: `Tasks: Run Task`.

## About DOM (HTMX, Shoelace) and Session

Resources:

- <https://shoelace.style>
- <https://htmx.org/docs/>
- <https://htmx.org/reference/>
- <https://pkg.go.dev/html/template>
- <https://docs.gofiber.io/next/> - v3, not v2!

We are using HTMX. JavaScript (TypeScript) is an utility for importing
libraries, extending DOM and web-components functionality. We are fetching HTML
from the server instead of JSON.

### About templates

Files in the [./web/templates](./web/templates) can be rendered through Go's
template language: <https://pkg.go.dev/html/template>.

That means, you can use specific syntax and replacements, but the variables
should be declared by the server. You can find more it in the server code base
(./internal). Specific functions are declared in the engine file
(./internal/engine.go).
