// deno-lint-ignore-file no-window-prefix no-window
import { findLastMessage, findLastMessageVisibleDate } from "./lib.ts";

window.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById(
        "send-message-form",
    ) as HTMLFormElement | undefined;
    if (!form) return;
    form.addEventListener(
        "htmx:oobAfterSwap",
        function (this: typeof form) {
            this.reset();
        },
    );
    findLastMessage()?.scrollIntoView();
    form.addEventListener("htmx:wsAfterMessage", function (event) {
        const message = (event as CustomEvent).detail.message as string;
        if (!message) return;

        mergeMessageJoinElement(message);
        form.reset();
    });
});

function mergeMessageJoinElement(message: string): void {
    // join same author messages
    const lastMessage = findLastMessageVisibleDate();

    if (!lastMessage) return;

    const lastMessageAuthorId = lastMessage.getAttribute("data-author");
    const newMessageAuthorId = message.match(
        /(?<=data-author=")\d+(?=")/,
    )?.[0];
    if (!lastMessageAuthorId || !newMessageAuthorId) {
        return;
    }
    const shouldJoin = lastMessageAuthorId === newMessageAuthorId;

    // join dates
    const lastMessageCreatedAt = lastMessage.getAttribute(
        "data-created-at",
    );
    const newMessageCreatedAt = message.match(
        /(?<=data-created-at=").+?(?=")/,
    )?.[0];
    if (!lastMessageCreatedAt || !newMessageCreatedAt) {
        return;
    }
    const dateDiff = new Date(lastMessageCreatedAt).getTime() -
        new Date(newMessageCreatedAt).getTime();
    const shouldJoinDate = dateDiff < 1000 * 60 * 5; // 5 minutes

    if (shouldJoin) {
        lastMessage.insertAdjacentHTML(
            "beforebegin",
            `<div class="join${shouldJoinDate ? " date" : ""}"></div>`,
        );
    }
}
