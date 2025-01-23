# Contributing

## Changing the code base

The best way is to use 2 terminals:

```bash
deno task serve
```

```bash
deno task watch
```

> [!NOTE]
> You can use VSC tasks for this.

The `watch` and `build` scripts are not tied with the server code base. That
means you should restart your server if you are making changes to the
`./internal` or if you are using `build`and making changes to the `./web`. Also
you should reload pages manually.

## About DOM (HTMX, Shoelace) and Session

Resources:

- <https://shoelace.style>
- <https://htmx.org/docs/>
- <https://htmx.org/reference/>
- <https://pkg.go.dev/html/template>
- <https://docs.gofiber.io/next/> - v3, not v2!

We are using HTMX. JavaScript (TypeScript) is an utility for importing
libraries, extending DOM and web-components functionality. We are fetching HTML
from the server instead of JSON - use the power of hypertext with HTMX.

> [!WARNING]
> DOM manipulations should be provided through HTMX and the server. Cookies
> should be changed by the server, if possible.
>
> Always send HTML as a response, if the request initialized by HTMX.

### About templates

Files in the [./web/templates](./web/templates) can be rendered through Go's
template language: <https://pkg.go.dev/html/template>.

That means, you can use specific syntax and replacements, but the variables
should be declared by the server, such as `{{.User}}`.

The User and other variables should be used to generate user/group-specific
content (logout button, profile, etc): `{{- if ne .User nil}}`,
`{{- if eq .User nil}}`.
