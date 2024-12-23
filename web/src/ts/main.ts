import "./theme.ts";
import { setBasePath } from "@shoelace-style/shoelace";

import * as HTMX from "htmx.org";
import "./shoelace-htmx-extension.js";
import "./shoelace-open-hash.ts";

declare namespace globalThis {
    let htmx: typeof HTMX;
}
globalThis.htmx = HTMX;

import("htmx-ext-debug");

setBasePath("/static/shoelace");
