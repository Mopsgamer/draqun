import "./theme.ts";
import { setBasePath } from "@shoelace-style/shoelace";

import htmx from "htmx.org";
import type HTMX from "htmx.org";
import "./shoelace-htmx-extension.ts";
import "./shoelace-open-hash.ts";
import { domLoaded, initAnchorHeadersFor } from "./lib.ts";

declare namespace globalThis {
    let htmx: typeof HTMX.default;
}
globalThis.htmx = htmx as unknown as typeof HTMX.default;

(htmx as unknown as typeof htmx.default).config.methodsThatUseUrlParams.length = 0

import("htmx-ext-debug");

setBasePath("/static/shoelace");

domLoaded.then(() => initAnchorHeadersFor(document.body));
