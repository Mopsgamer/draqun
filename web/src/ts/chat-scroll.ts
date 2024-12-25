import { findLastMessage } from "./lib.ts";

const chat = document.getElementById("chat")!;

findLastMessage()?.scrollIntoView();
document.addEventListener("htmx:wsAfterMessage", function (event) {
    if ((event as CustomEvent).detail.message == "") {
        return;
    }
    // scroll down
    const isScrolledToBottom =
        chat.scrollHeight - chat.scrollTop - chat.clientHeight < 450;
    if (isScrolledToBottom) {
        chat.scrollTop = chat.scrollHeight;
    }
});
