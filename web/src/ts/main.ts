import { setBasePath } from "@shoelace-style/shoelace";

import * as HTMX from "htmx.org"
import "./shoelace-htmx-extension.js";
import "./shoelace-dialog-from-hash.ts";
import "./chat-scroll.ts";
import "./chat-send.ts";

declare namespace globalThis {
    let htmx: typeof HTMX
}
globalThis.htmx = HTMX

import("htmx-ext-debug");
import("htmx-ext-ws");

setBasePath("/static/shoelace");
