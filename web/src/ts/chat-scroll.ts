// deno-lint-ignore-file no-window-prefix no-window
import { findLastMessage } from "./lib.ts";

window.addEventListener("DOMContentLoaded", () => {
    findLastMessage()?.scrollIntoView();
    document.addEventListener("htmx:wsAfterMessage", function () {
        // scroll down
        const chat = document.getElementById("chat");
        if (!chat) return;
        const isScrolledToBottom =
            chat.scrollHeight - chat.scrollTop - chat.clientHeight < 450;
        if (isScrolledToBottom) {
            chat.scrollTop = chat.scrollHeight;
        }
    });
});
