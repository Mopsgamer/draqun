# Contributing

## First setup

1. Install required tools.
   - MySQL@^8.0
     - [Windows installation](https://winstall.app/apps/Oracle.MySQL),
       [Ubuntu installation](https://documentation.ubuntu.com/server/how-to/databases/install-mysql/index.html),
       [Mac installation](https://dev.mysql.com/doc/refman/8.4/en/macos-installation-pkg.html)
     - Recommended db name: `mysql`.
     - Recommended user: `admin`.
   - Go@^1.24 ([Installation](https://go.dev/doc/install))
   - Deno@^2.4 ([Installation](https://deno.com/))
2. [Fork](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/fork-a-repo)
   and
   [clone](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository)
   the repository.
3. Open it in your favorite editor. [VSC](https://code.visualstudio.com/) is
   recommended. For small changes you can also use
   [in-browser GitHub editor](https://docs.github.com/en/codespaces/the-githubdev-web-based-editor).
4. Open terminal. You can use built-in
   [VSC terminal](https://code.visualstudio.com/docs/terminal/getting-started).
5. Run `deno install --allow-scripts` to install required client dependencies.
6. Run `deno task init` to create `.env` file and initialize DB (use `deno task init nodb` to skip db initialization).
7. Run `deno task compile:client` to create client files.
8. Change the `.env` file.
   - Set up server connection with MySQL.
   - Set up JWT secret.
9. Run `deno task serve` to start the server.

## Compilation

Compilation is used to create a standalone server binary. It is useful for
deploying the server to production or for distributing it as a standalone
application.

Available go build tags:

- Environment:
  - `[none]` enables client files watching.
  - `prod` normal mode.
- Client embedding:
  - `[none]` enables client files embedding. The server binary will become
    standalone.
  - `lite` disables files embedding. The server binary will use closest
    ./client/static and ./client/templates directories. This option makes the
    server binary more flexible and reduces its size.

Example: `go -o dist/server.exe -tags lite,prod .`

Available deno tasks:

```bash
deno task compile:server # prod
deno task compile:server dev # lite
deno task compile:server:cross # prod
deno task compile:server:cross dev #lite
```

## Changing the code base

Client code base (./client) is not tied with the server code base (./server).
The best way is to use 2 terminals (3-rd for other tasks):

> [!NOTE]
> You can use Visual Studio Code's task commands: `Tasks: Run Task`.
>
> - Compile Client & Watch
> - Serve

```bash
deno task compile:client watch
```

```bash
deno task serve
```

> [!WARNING]
> The `serve` script can ignore new files, so it should be started after
> `compile:client` script generates all files. If you are using `watch`, wait
> for "watching..." message.

## How to write commit messages and PR names.

We use [Conventional Commit messages](https://www.conventionalcommits.org/) to
automate version management.

Most common commit message prefixes are:

- `fix:` which represents bug fixes and generate a patch release.
- `feat:` which represents a new feature and generate a minor release.
- `impr:` which represents an improvement and generate a minor release.
- `chore:` which represents a development environment change and generate a
  patch release.
- `docs:` which represents documentation change and generate a patch release.
- `style:` which represents a code style change and generate a patch release.
- `test:` which represents a test change and generate a patch release.
- `BREAKING CHANGE:` which represents a breaking change and generate a major
  release. Or you are able to use `!` at the end of the prefix. For example
  `feat!: new feature` or `fix!: bug fix`.
- Use `prefix(module):` or `prefix(module)!:` to specify a module. For example,
  `feat(auth): new login page` or `fix(auth)!: login page on mobile devices`.

Messages itself should be lowercase, without punctuation at the end and should
be short, but descriptive.

## Releasing

> [!NOTE]
> You should be a repository owner or have write access to create a release.

You can create new releases and git tag automatically based on commits or custom
release type, using GitHub workflow manual execution (dispatch). Available
options:

- `keep`: do not increment;
- `from-commits`: determine from commit messages;
- `patch`: 1.2.0 → 1.2.1 → 1.2.2;
- `minor`: 1.2.0 → 1.3.0 → 1.4.0;
- `major`: 1.2.0 → 2.0.0 → 3.0.0;
- `prepatch`: 1.2.0 → 1.2.1-0 → 1.2.2-1;
- `preminor`: 1.2.0 → 1.3.1-0 → 1.4.0-1;
- `premajor`: 1.2.0 → 2.0.0-0 → 3.0.0-1;
- `pre`: 1.2.0 → 1.2.0-0 → 1.2.0-1;
- `prerelease`: 1.2.0 → 1.2.1-0 → 1.2.1-1;

You can also use deno task to create a release from your machine, but it is not
recommended:

```bash
deno task release
```

You can get next version and changelog output without creating a release:

```bash
deno task release --dry-run
```

## About DOM (HTMX, Shoelace) and Session

Resources:

- <https://shoelace.style>
- <https://htmx.org/docs/>
- <https://htmx.org/reference/>
- <https://pkg.go.dev/html/template>
- <https://docs.gofiber.io/next/> - v3 (Next), not v2!

We are using HTMX. JavaScript (TypeScript) is an utility for importing
libraries, extending DOM and web-components functionality. We are fetching HTML
from the server instead of JSON.

The session stored in cookies and should be changed this way:

1. Client sends request to change own cookies.
2. Server responds with new cookies.

### About templates

Files in the [./client/templates](./client/templates) can be rendered through
Go's template language: <https://pkg.go.dev/html/template>.

That means, you can use specific syntax and replacements, but the variables
should be declared by the server. You can find more it in the
[./server](./server). Specific functions are declared in the
[./server/engine.go](./server/engine.go). .
