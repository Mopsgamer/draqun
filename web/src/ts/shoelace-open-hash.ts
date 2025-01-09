import type { SlDialog, SlDrawer } from "@shoelace-style/shoelace";

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
    const selector = "sl-dialog, sl-drawer";
    type SlOpenable = SlDialog | SlDrawer;
    const slOpenableList = document.querySelectorAll<SlOpenable>(
        selector,
    );
    for (const slOpenable of slOpenableList) {
        if (
            slOpenable.id !== id &&
            !slOpenable.querySelector("#" + id)
        ) {
            slOpenable.open = false;
            continue;
        }

        foundDialogFromHash = true;
        slOpenable.open = true;
        slOpenable.addEventListener("sl-after-hide", () => {
            if (!location.hash.includes(slOpenable.id)) return;

            const anotherOpened = document.querySelector(
                `:is(${selector})[open]:not([open=false])`,
            );
            location.hash = anotherOpened?.id ?? "";
            cleanHash();
        }, { once: true });
    }

    if (!foundDialogFromHash) {
        cleanHash();
    }
}

addEventListener("hashchange", () => openDialogFromHash());
addEventListener("load", function () {
    openDialogFromHash();
});
