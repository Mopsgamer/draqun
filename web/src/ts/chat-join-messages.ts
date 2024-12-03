// deno-lint-ignore-file no-window-prefix no-window
import { chatJoinMessages } from "./lib.ts";

window.addEventListener("DOMContentLoaded", () => {
    const chat = document.getElementById("chat");
    if (!chat) return;

    const observer = new MutationObserver(() => {
        chatJoinMessages();
    });
    observer.observe(chat, { childList: true, subtree: true });
});
