# Draqun Architecture & Agent Guide

Compact reference for implementing channels, permissions, WebSockets, UI layouts, and security.

Keep this file updated.

## 1. Tooling & Commands

- **Stack**: Go (backend / Fiber), SQLite + Goqu (DB), Bun (frontend compiler), Tailwind + DaisyUI, HTMX.
- **Commands**:
  - `bun run front`: Compile frontend assets.
  - `go test ./server/...`: Run backend tests.
  - `go mod tidy`: Clean Go dependencies.
- **MCP / Testing**:
  - Prefer `agent-browser` CLI (`bun x agent-browser`) for frontend checks.
- **Directory Structure & Architecture**:
  - `server/model/`: DB structs, Goqu queries, and SQL schemas (`server/model/queries/`). Avoid importing `server/model` inside `server/session` to prevent Go package import cycles.
  - `server/session/`: Realtime state hubs (WebSockets, in-memory collab connections). Keep lightweight with primitive types.
  - `server/perms/`: Auth parsing, RBAC validation, and route middlewares.
  - `server/routes/`: Route declarations and page/api controllers (`groups.go`, `pages-chat.go`, etc.).
  - `server/htmx/` & `server/render/`: HTMX helper alerts, templates, and OOB swap renderers.
  - `client/templates/`: Go HTML templates (`partials/` for component chunks, root files for pages).
  - `client/src/`: Raw Tailwind CSS and browser-only TypeScript (only for DOM helpers/events, never for business/session state).
- **Rules**:
  - No backward compatibility needed.
  - Do not bump `package.json` version or alter proxy configs unless requested.
  - **Server-Side First**: Prefer server-side rendering with Go Fiber + HTMX for all UI interactions and state. Do not create client-side TypeScript managers.

## 2. Layout & UI Structure (`client/templates/`)

### Layout Anatomy (`partials/app-group-header.tmpl`, `partials/app-group.tmpl`)

- **Header (`#header` / `header.navbar`)**:
  - `navbar-start`: Logo (`/app`), sidebar toggle (`.group-channels-toggle`), active channel moniker trigger (`.group-members-toggle`).
  - `navbar-end`: Collaboration connection & input/output controls (`partials/collab-connection-status`), notifications (`partials/notification-bell`), user menu (`partials/avatar-context-menu`).
- **Left Column (`#channels-view` / `.secondary-view`)**:
  - Group info (avatar, moniker, `@name`), settings link (`/app/groups/:id/settings`).
  - Channel list (`#channels`): Top-level channels + subchannels/threads. Active channel highlighted (`bg-primary/15 text-primary`).
  - Filtered by `GetUserPermissionLevel`: Channels with `PermLevelDisable` on `read` are excluded.
- **Center Column (`#chat-view`)**:
  - Main channel workspace. Conditionals select the view:
    - Ticket container: `partials/ticket-list`
    - Ticket subchannel: `partials/ticket-detail`
    - Otel dashboard: `partials/otel-dashboard`
    - Server config: `partials/config-dashboard`
    - Wall: `partials/wall-view`
    - Collab: Triggers header connection session without swapping main workspace view.
    - Text/default: `<main id="chat">` with `#chat-messages-:id` + `partials/chat-input`
- **Right Column (`#members-view` / `.secondary-view`)**:
  - Active channel moniker/name, star button (`hx-post="/groups/:id/channels/:id/star"`), close toggle.
  - Channel markdown description (if any).
  - Member count (`#members-count`) and paginated member list (`#members`).

### DaisyUI & Dialog Rules

- Use DaisyUI components (`btn`, `btn-primary`, `btn-ghost`, `avatar`, `dropdown`, `menu`, etc.).
- Extract repetitive utility patterns into CSS classes (`client/src/tailwindcss/components.css`).
- **Popups & Menus**:
  - Use popover API.
  - **Forbidden**: Native browser dialogs (`alert`, `confirm`, `prompt`) and `hx-confirm`. Use DaisyUI `<details class="dropdown">`, modals, or inline buttons.

## 3. Adding New Channel Types

1. **Model** (`server/model/channel.go`):
   - Add constant: `const ChannelTypeNew ChannelType = "..."`.
   - Update `(ct ChannelType) IsValid() bool`.
   - Add helper: `func (c Channel) IsNew() bool { return c.Type == ChannelTypeNew }`.
   - Update `server/model/queries/create_channels.sql` if applicable.
2. **Icon & Navigation**:
   - Map icon in `client/templates/partials/channel-icon-class.tmpl`.
3. **Dispatch & Templates**:
   - In `server/routes/pages-chat.go` (`handleGroupChat`), set view flag (e.g. `IsNew: activeChannel.IsNew()`).
   - In `client/templates/partials/app-group.tmpl`, add branch in `#chat-view` to render `partials/new-view`.
   - Create `client/templates/partials/new-view.tmpl`.
4. **Channel Routes**:
   - Add specific handlers in `server/routes/groups.go` under `/groups/:group_id/channels/:channel_id/...`.

## 4. Permissions System (`server/perms/`, `server/model/permission.go`)

### Concepts

- **`PermName`**: `read`, `write`, `delete`, `members`, `roles`, `settings`, `admin`.
- **`PermLevel`**:
  - `all`: Unrestricted access.
  - `own`: Limited to author or assignee. **Forbidden** for `admin`, `roles`, `settings`.
  - `disable`: Blocked.
- **1 permission per perm type** (not bundled; sorted by `priority DESC, id DESC`). Priority is managed via reordering controls, not direct form edits.

### Evaluation (`GetUserPermissionLevel`)

1. Group Owner -> `PermLevelAll`.
2. Admin `PermLevelAll` -> grants `PermLevelAll` for other permissions.
3. Check rules in `app_perms` ordered by `priority DESC, id DESC`:
   - Match user IDs and role IDs (including `@everyone`).
   - Match channel IDs (subchannels inherit parent channel rules).
4. Subchannel/ticket with level `own`: resolves to `own` only if user is `AuthorId` or assignee (`IsChannelAssignee`). Otherwise `disable`.
5. Default fallbacks: `read`/`write`/`delete`/`members` -> `own`; `roles`/`settings`/`admin` -> `disable`.

### Enforcing Permissions

- Middleware: `perms.RequirePermLevel(model.PermNameWrite, model.PermLevelAll)` or `perms.RequirePerm("group_id", model.PermNameRead)`.
  - If `channel_id` is 0 in nested subroutes, fallback to `parent_id`.
- Handlers: Call `model.GetUserPermissionLevel(groupId, userId, channelId, permName)`.
- Query params in Fiber GET handlers: Use fallbacks (`ctx.Query` or `strconv.ParseUint`).

## 5. WebSockets & Realtime (`server/session/`)

- **Route**: `GET /ws/:group_id/` in `server/routes/groups.go` (requires `PermNameRead`).
- **Collaboration rooms** (`server/session/collab.go`, `server/routes/groups.go`):
  - Collaboration subchannels are virtual room numbers, not channel names. Accept only positive numeric room IDs; do not treat arbitrary strings (for example, `"g"`) as room IDs.
  - The join route allocates the next available room number for `subchannel=new` and validates explicitly requested room numbers. Keep the active-room list limited to valid numeric rooms.
  - Signaling must use the user's actual joined room from `CollabHub`, not trust a client-supplied subchannel. RTP relay matching must use that authoritative room (as well as group and channel IDs).
  - Room membership is in memory; restarting the server clears stale room state.
- **Client**: `hx-ext="ws"` on `#ws-chat` in `client/templates/partials/app-group.tmpl`:
  ```html
  <div
  	id="ws-chat"
  	class="collapse"
  	hx-ext="ws"
  	ws-connect="/groups/ws/{{.Group.Id}}"
  	hx-trigger="every 1.6s"
  	ws-send=""
  ></div>
  ```
- **Session Hub (`server/session/connections.go`)**:
  - Manage: `session.UserSessionMap.Connect(userId, ws)` / `.Close(userId, ws)`.
  - Send: `session.UserSessionMap.PushUser(userId, htmlPayload)` / `.PushGroup(groupId, htmlPayload)`.
- **Broadcasting via OOB**:
  - Render template: `render.RenderBuffer(app, "partials/...", data)`.
  - Wrap OOB swap: `render.WrapOob("innerHTML:#target-id", &renderedHtml)`.
  - Push to group/user.

## 6. Security Checklist

- **Suspended Users**: Check `user.IsSuspended` in auth / token middleware (`server/perms/load.go`). Block reports (`/reports/*`) and actions from suspended users.
- **Global IP Blocking**: Enforced only via global middleware in `server/app.go` (`AlertIPBlocked`). Never duplicate in route handlers.
- **Input & Parameter Checks**:
  - Message create (`POST /:group_id/messages/create`): Validate write perm explicitly against body `request.ChannelId`.
  - Member search (`handleMemberSearch`): Validate read perm when `channel_id` is specified; return 403 (`AlertGroupMemberNotAllowed`) if forbidden.
- **Admin Member Actions**: Kicking/banning requires `PermLevelAll` for `members`.
- **Otel Channels**: `ChannelTypeOtel` restricted to global admins (`IsDraqunAdminOrMod`).
- **Sliding JWT Sessions**: `UserByAuthFromCtx` refreshes token cookie when remaining lifespan < 50%.

## 7. Key File Map

| Area                 | Key Files                                                                             |
| -------------------- | ------------------------------------------------------------------------------------- |
| Channels             | `server/model/channel.go`, `server/model/queries/create_channels.sql`                 |
| Permissions          | `server/model/permission.go`, `server/perms/middleware.go`, `server/perms/load.go`    |
| WebSockets           | `server/session/controller_ws.go`, `server/session/connections.go`                    |
| Group Chat & Routes  | `server/routes/pages-chat.go`, `server/routes/groups.go`                              |
| App Shell & Partials | `client/templates/partials/app-group.tmpl`, `partials/app-group-header.tmpl`          |
| Collab Connections   | `server/session/collab.go`, `client/templates/partials/collab-connection-status.tmpl` |
| Rendering & HTMX     | `server/render/render.go`, `server/htmx/`                                             |
| Styles               | `client/src/tailwindcss/components.css`                                               |
