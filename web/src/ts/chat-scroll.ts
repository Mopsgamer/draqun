// deno-lint-ignore-file no-window-prefix no-window
import { findLastMessage } from "./lib.ts";

window.addEventListener("DOMContentLoaded", () => {
    const chat = document.getElementById("chat");
    if (!chat) return;
    findLastMessage()?.scrollIntoView();
    document.addEventListener("htmx:wsAfterMessage", function (event) {
        if((event as CustomEvent).detail.message == "") {
            return
        }
        // scroll down
        const isScrolledToBottom =
            chat.scrollHeight - chat.scrollTop - chat.clientHeight < 450;
        if (isScrolledToBottom) {
            chat.scrollTop = chat.scrollHeight;
        }
    });
});
