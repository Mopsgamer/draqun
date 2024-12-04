import { setBasePath } from "@shoelace-style/shoelace";

import * as HTMX from "htmx.org";
import "./shoelace-htmx-extension.js";
import "./shoelace-dialog-from-hash.ts";

declare namespace globalThis {
    let htmx: typeof HTMX;
}
globalThis.htmx = HTMX;

import("htmx-ext-debug");

setBasePath("/static/shoelace");
