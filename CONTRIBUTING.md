# Contributing to restapp

## About DOM (HTMX, Shoelace) and Session

Resources:

- <https://shoelace.style>
- <https://htmx.org/docs/>
- <https://htmx.org/reference/>
- <https://pkg.go.dev/html/template>
- <https://docs.gofiber.io/next/> - Should be NEXT! We use v3 not v2!

We are using HTMX. That means we are using JS as an utility for importing
libraries and extending DOM and web-components functionality (actually we are
using TS). We are fetching HTML from the server instead of JSON - use the power
of hypertext with HTMX.

TS should NOT be used, if possible. Possible reasons to add :

- Importing new standalone library: HTMX, tailwind, any web-components library,
  etc.
- Fixing or extending DOM features: make libraries to work together, adding new
  web-component, etc.

> [!WARNING]
> DOM manipulations should be provided through HTMX and the server. Cookies
> should be changed by the server, if possible.
>
> Always send HTML as a response, if the request initialized by HTMX. Script
> tags available.

### About templates

Files in web/templates can be rendered through Go's template language:
<https://pkg.go.dev/html/template>

This means, you can use specific syntax and replacements, but the variables
should be declared by the server, such as `{{.User}}`.

The User variable should be used to generate user-specific content (logout
button, profile, etc): `{{if ne .User nil}}`, `{{if eq .User nil}}`.
