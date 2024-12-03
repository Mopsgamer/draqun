# Contributing to restapp

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

Possible reasons to add the TypeScript:

- Importing or fixing a library or plugin: HTMX, tailwind, any web-components
  library, etc.
- Fixing or extending DOM features: make libraries to work together, adding new
  web-component, etc.

> [!WARNING]
> DOM manipulations should be provided through HTMX and the server. Cookies
> should be changed by the server, if possible.
>
> Always send HTML as a response, if the request initialized by HTMX. Script
> tags available.

### About templates

Files in the [./web/templates](./web/templates) can be rendered through Go's
template language: <https://pkg.go.dev/html/template>.

That means, you can use specific syntax and replacements, but the variables
should be declared by the server, such as `{{.User}}`.

The User and other variables should be used to generate user/group-specific
content (logout button, profile, etc): `{{- if ne .User nil}}`,
`{{- if eq .User nil}}`.
