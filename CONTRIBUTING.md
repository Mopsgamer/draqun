# Contributing

## First setup

1. Install required tools.
   - MySQL `8||9`.
     - Recommended db name: `mysql`.
     - Recommended user: `admin`.
   - [Go](https://go.dev/doc/install) `^1.25`
   - [Deno](https://deno.com/) `^2.5`
2. [Fork](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/fork-a-repo)
   and
   [clone](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository)
   the repository.
3. Open it in your favorite editor. [VSC](https://code.visualstudio.com/) is
   recommended. For small changes you can also use
   [web-based GitHub editor](https://docs.github.com/en/codespaces/the-githubdev-web-based-editor).
4. Open terminal. You can use built-in
   [VSC terminal](https://code.visualstudio.com/docs/terminal/getting-started).
5. Run `deno install --allow-scripts` to install required client dependencies.
6. Run `go mod download` to install required server dependencies (optional).
7. Run `deno task init` to create `.env` file and initialize the database.
   `noenv` and `nodb` options are available.
8. Run `deno task compile:client` to create client files.
9. Change the `.env` file.
   - Set up server connection with MySQL.
   - Set up JWT secret.
10. Run `deno task dev` to start the server.

## Making changes

The best way is to use two terminals, with a third for other tasks:

> [!NOTE]
> You can use Visual Studio Code's task commands: `Tasks: Run Task`.
>
> - Compile Client & Watch
> - Serve

```bash
deno task compile:client watch
```

```bash
deno task dev
```

Before committing your changes, make sure to run the prepare task to format:

```bash
deno task prepare
```

### Resources

- <https://shoelace.style>
- <https://htmx.org/docs/>
- <https://htmx.org/reference/>
- <https://pkg.go.dev/html/template>
- <https://docs.gofiber.io/next/>

## How to write commit messages and PR names.

We use [Conventional Commit messages](https://www.conventionalcommits.org/) to
automate version management.

Most common commit message prefixes are:

- `fix:` which represents bug fixes and generate a patch release.
- `feat:` which represents a new feature and generate a minor release.
- `impr:` which represents an improvement and generate a minor release.
- `ci:` which represents a CI/CD change and generate a patch release.
- `docs:` which represents documentation change and generate a patch release.
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

## About releases

> [!NOTE]
> You should be a repository owner or have write access to create a release.

You can create new release and git tag automatically using GitHub workflow.

Available options:

- `patch`
- `minor`
- `major`
- `release`
- `alpha patch`
- `alpha minor`
- `alpha major`
- `alpha release`
- `beta patch`
- `beta minor`
- `beta major`
- `beta release`

Examples:

| Release type  | Result                        |
| ------------- | ----------------------------- |
| major         | 0.0.1 → 1.0.0 → 2.0.0         |
| minor         | 0.0.1 → 0.1.0 → 0.2.0         |
| patch         | 0.0.1 → 0.0.2 → 0.0.3         |
| patch         | 0.0.1.alpha.1 → 0.0.1         |
| patch alpha   | 0.0.1.alpha.1 → 0.0.2-alpha.0 |
| patch beta    | 0.0.1.alpha.1 → 0.0.2-beta.0  |
| release       | 0.0.1.alpha.1 → 0.0.1         |
| release alpha | 0.0.1.alpha.1 → 0.0.1-alpha.2 |
| release beta  | 0.0.1.alpha.1 → 0.0.1-beta.0  |
| alpha         | 0.0.1.alpha.1 → 1.0.0-alpha.0 |
| beta          | 0.0.1.alpha.1 → 1.0.0-beta.0  |

You can get next version and changelog output without creating a release:

```bash
deno task release
```

You can also create a release from your machine, but it is not recommended:

```bash
deno task release --force
```

## Compilation

Creating a standalone server binary is useful for deploying the server to
production or for distributing it as a standalone application.

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
# -tags prod
deno task compile:server
deno task compile:server:cross

# -tags lite
deno task compile:server dev
deno task compile:server:cross dev
```
