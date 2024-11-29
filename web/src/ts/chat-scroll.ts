// deno-lint-ignore-file
window.addEventListener("DOMContentLoaded", () => {
    const chat = document.getElementById("chat");
    if (!chat) return;
    function findLastMessage(): Element | undefined {
        return Array.from(chat!.children).reverse().find(
            (c) => c.classList.contains("message"),
        );
    }
    findLastMessage()?.scrollIntoView();
    document.addEventListener("htmx:wsAfterMessage", function (event: any) {
        const message = event.detail.message as string;
        if (!message) return;

        // join same author messages
        const lastMessage = findLastMessage();

        if (!lastMessage) return;

        const lastMessageAuthorId = lastMessage.getAttribute("data-author");
        const newMessageAuthorId = message.match(
            /(?<=data-author=")\d+(?=")/,
        )?.[0];
        if (!lastMessageAuthorId || !newMessageAuthorId) {
            return console.error("can not join messages (author)");
        }
        const shouldJoin = lastMessageAuthorId === newMessageAuthorId;

        const lastMessageCreatedAt = lastMessage.getAttribute(
            "data-created-at",
        );
        const newMessageCreatedAt = message.match(
            /(?<=data-created-at=").+?(?=")/,
        )?.[0];
        if (!lastMessageCreatedAt || !newMessageCreatedAt) {
            return console.error("can not join messages (author+date)");
        }
        const dateDiff = Number(new Date(lastMessageCreatedAt)) -
            Number(new Date(newMessageCreatedAt));
        console.log(new Date(lastMessageCreatedAt));
        const shouldJoinDate = dateDiff < 1e3 * 60 * 5; // 5 minutes

        if (shouldJoin) {
            lastMessage.insertAdjacentHTML(
                "beforebegin",
                `<div class="join${shouldJoinDate ? " date" : ""}"></div>`,
            );
        }

        // scroll down
        console.log();
        const isScrolledToBottom =
            chat.scrollHeight - chat.scrollTop - chat.clientHeight < 450;
        if (isScrolledToBottom) {
            chat.scrollTop = chat.scrollHeight;
        }
    });
});
