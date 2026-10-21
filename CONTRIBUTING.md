# Contributing

## First setup

1. Install required tools.
   - Git, obviously
   - [Go](https://go.dev/doc/install) `^1.27.3`
   - [Bun](https://bun.sh/) `^1.2.0`
2. [Fork](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/fork-a-repo)
   and
   [clone](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository)
   the fork.
3. Open it in your favorite editor. [VSC](https://code.visualstudio.com/) is
   recommended. You can also use
   [web-based GitHub editor](https://docs.github.com/en/codespaces/the-githubdev-web-based-editor)
   or a [Codespace](https://docs.github.com/en/codespaces/quickstart).
4. Open terminal. You can use built-in
   [VSC terminal](https://code.visualstudio.com/docs/terminal/getting-started).
5. Run `bun install` to install required client dependencies.
6. Run `go mod download` to install required server dependencies (optional).
7. Environment (`.env`) creation, database table creation, and schema migrations
   are executed automatically by the Go backend server upon startup. You can
   customize `.env` values (such as `JWT_KEY` or `PORT`) as needed.
8. Run `bun run dev` to build client files and start the server.
9. Open your browser and navigate to `http://localhost:3000` (or hold the
   <kbd>Ctrl</kbd> key and click the link in the terminal).

## Making changes

Run `bun run dev` to build and serve locally. It also watches client files and
rebuilds frontend assets when they change:

> [!NOTE]
> You can use Visual Studio Code's task commands: `Tasks: Run Task`.
>
> - `bun run dev` to build frontend assets, watch for client changes, and run
>   the development server.

```bash
bun run dev
```

Other available VS Code tasks mirror the Bun tasks: `bun lint`, `bun fmt`, `bun run back`, and `bun run front` for one-time frontend builds.

### Resources

- <https://daisyui.com>
- <https://htmx.org/docs/>
- <https://htmx.org/reference/>
- <https://pkg.go.dev/html/template>
- <https://docs.gofiber.io/next/>

## How to write commit messages and PR names

Write clear, concise PR titles that explain the "why" behind changes.

We use [Conventional Commits](https://www.conventionalcommits.org/). Since we
squash on merge, the PR title becomes the commit message in `main`, so it's
important to get it right.

Format: `type(scope): description`

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`,
`ci`, `chore`, `revert`

**Scopes (optional):** `a11y`, `blog`, `deps`, `docs`, `i18n`, `ui`

**Examples:**

- `fix: resolve search pagination issue`
- `feat: add package version comparison`
- `fix(i18n): update French translations`
- `chore(deps): update esbuild to v0.8`

Where front end changes are made, please include before and after screenshots in
your pull request description.

> [!NOTE]
> Use lowercase letters in your pull request title. Individual commit messages
> within your PR don't need to follow this format since they'll be squashed.

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
| alpha patch   | 0.0.1 → 0.0.2-alpha.0         |
| alpha minor   | 0.0.1 → 0.1.0-alpha.0         |
| alpha major   | 0.0.1 → 1.0.0-alpha.0         |
| beta patch    | 0.0.1 → 0.0.2-beta.0          |
| beta minor    | 0.0.1 → 0.1.0-beta.0          |
| beta major    | 0.0.1 → 1.0.0-beta.0          |
| patch         | 0.0.1-alpha.1 → 0.0.1         |
| release       | 0.0.1-alpha.1 → 0.0.1         |
| release alpha | 0.0.1-alpha.1 → 0.0.1-alpha.2 |
| release beta  | 0.0.1-alpha.1 → 0.0.1-beta.0  |
| patch alpha   | 0.0.1-alpha.1 → 0.0.2-alpha.0 |
| patch beta    | 0.0.1-alpha.1 → 0.0.2-beta.0  |
| alpha         | 0.0.1-alpha.1 → 1.0.0-alpha.0 |
| beta          | 0.0.1-alpha.1 → 1.0.0-beta.0  |

You can get next version and changelog output without creating a release:

```bash
bun run release
```

You can also create a release from your machine, but it is not recommended:

```bash
bun run release --force
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

Available Bun tasks:

```bash
# -tags prod
bun run back
bun run back:cross

# -tags lite
bun run back dev
bun run back:cross dev
```

## Burn it all down

### Purge global caches & artifacts

Even after deleting the folder, Bun and Go leave cached modules and build
artifacts in your system directories. Run these commands to wipe them out
completely from your local environment:

```bash
# Remove cached modules, lockfiles, etc.
bun pm cache rm
# Remove all build and module download caches
go clean -cache -modcache
```

- Bun Registry Cache: Located at `~/.bun/install/cache` (Linux/macOS).
- Go Module Cache: Located at `$GOPATH/pkg/mod` (usually `~/go/pkg/mod`)
  (Linux/macOS) or `%USERPROFILE%\go\pkg\mod` (Windows).
