import type { SlDialog } from "@shoelace-style/shoelace";

function removeHash() {
    let scrollV, scrollH;

    if ("pushState" in history) {
        history.pushState(
            "",
            document.title,
            location.pathname + location.search,
        );
    } else {
        // Prevent scrolling by storing the page's current scroll offset
        scrollV = document.body.scrollTop;
        scrollH = document.body.scrollLeft;

        location.hash = "";

        // Restore the scroll offset, should be flicker free
        document.body.scrollTop = scrollV;
        document.body.scrollLeft = scrollH;
    }
}

function cleanHash() {
    if (location.hash !== "") {
        return;
    }
    removeHash();
}

function openDialogFromHash() {
    const id = /(?<=#)[a-zA-Z\d_-]+/.exec(location.hash)?.[0];
    if (!id) {
        cleanHash();
        return;
    }

    let foundDialogFromHash = false;
    for (const slDialog of document.querySelectorAll<SlDialog>("sl-dialog")) {
        if (slDialog.id === id || slDialog.querySelector("#" + id)) {
            foundDialogFromHash = true;
            slDialog.open = true;
            slDialog.addEventListener("sl-after-hide", () => {
                if (!location.hash.includes(slDialog.id)) {
                    return;
                }
                const anotherOpened = document.querySelector(
                    "sl-dialog[open]:not([open=false])",
                );
                location.hash = anotherOpened?.id ?? "";
                cleanHash();
            }, { once: true });
            continue;
        }
        slDialog.open = false;
    }

    if (!foundDialogFromHash) {
        cleanHash();
    }
}

addEventListener("hashchange", () => openDialogFromHash());
addEventListener("load", function () {
    openDialogFromHash();
});
