// deno-lint-ignore-file no-window-prefix no-window
import { findLastMessage } from "./lib.ts";

window.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById(
        "send-message-form",
    ) as HTMLFormElement | undefined;
    if (!form) return;

    form.addEventListener(
        "htmx:afterRequest",
        function (this: typeof form) {
            this.reset();
        },
    );
    findLastMessage()?.scrollIntoView();
});
