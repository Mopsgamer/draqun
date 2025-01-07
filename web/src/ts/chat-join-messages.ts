import { chatJoinMessages, domLoaded } from "./lib.ts";

domLoaded.then(() => {
    const chat = document.getElementById("chat");

    if (!chat) {
        return
    }

    const observer = new MutationObserver(() => {
        chatJoinMessages();
    });
    observer.observe(chat, { childList: true, subtree: true });
});
