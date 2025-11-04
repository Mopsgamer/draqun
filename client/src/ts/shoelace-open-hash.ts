import { SlDialog, type SlDrawer } from "@shoelace-style/shoelace";
import { domLoaded } from "./lib.ts";

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

    if (!openDialog(id)) {
        cleanHash();
    }
}

function openDialog(id: string) {
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

    return foundDialogFromHash;
}

document.addEventListener("hashchange", () => openDialogFromHash());
domLoaded.then(() => openDialogFromHash());
document.addEventListener("click", (e) => {
    if (!(e.target instanceof Element)) return;
    const a = e.target.closest(
        ':is(a, sl-button, sl-sl-icon-button)[href^="#"]',
    );
    if (!a) return;

    const dialogId = a.getAttribute("href")!.slice(1);
    openDialog(dialogId);
});
