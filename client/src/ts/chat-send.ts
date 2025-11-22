import { domLoaded, findLastMessage } from "./lib.ts";

domLoaded.then(() => {
    const form = document.getElementById(
        "send-message-form",
    ) as HTMLFormElement | null;

    if (!form) return;

    form.addEventListener(
        "htmx:afterRequest",
        () => {
            form.reset();
        },
    );

    findLastMessage()?.scrollIntoView();
});
