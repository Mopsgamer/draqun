import htmx from "htmx.org";
import type {
    SlDialog,
    SlInput,
    SlRadioGroup,
    SlRating,
} from "@shoelace-style/shoelace";

function onEvent(name: htmx.HtmxEvent, evt: CustomEvent) { // htmx type definitions sucks
    if (name !== "htmx:configRequest") {
        return;
    }

    if (evt.detail.elt.tagName !== "FORM") {
        return;
    }

    const form = evt.detail.elt as HTMLFormElement; // sucks

    const slElementList = Array.from(
        form.querySelectorAll("sl-radio-group, sl-rating, sl-input"),
    ) as Array<(SlRadioGroup | SlRating | SlInput) & HTMLFormElement>;

    for (const slElement of slElementList) {
        const isDisabled = !slElement.name || slElement.disabled ||
            slElement.closest("[disabled]");
        if (isDisabled) continue;

        const ratingOrInputName = (slElement.tagName === "SL-RATING" ||
            slElement.tagName === "SL-INPUT") &&
            slElement.getAttribute("name");

        let name = slElement.name;
        const value = slElement.value;
        if (ratingOrInputName) {
            name = ratingOrInputName;
        }

        evt.detail.parameters[name] = value;
    }
    // Prevent form submission if one or more fields are invalid.
    // form is always a form as per the main if statement
    if (!form.checkValidity()) {
        return false;
    }
}

htmx.defineExtension("shoelace", { onEvent });

for (
    const slDialog of Array.from(
        document.querySelectorAll("sl-dialog"),
    ) as SlDialog[]
) {
    openDialogIfHash(slDialog);
    slDialog.addEventListener("sl-hide", () => removeHash());
}

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

removeHash();

function openDialogIfHash(dialog: SlDialog): void {
    if (location.hash.startsWith(dialog.id)) {
        console.log("opening %o", dialog);
        dialog.open = true;
    }
}

function openDialogFromHash() {
    const hash = location.hash;
    const id = /(?<=#)[a-zA-Z\d_-]+/.exec(hash)?.[0];
    if (!id) {
        return;
    }
    const newDialog = document.getElementById(id) as SlDialog | null;
    if (!newDialog) {
        console.log(id);
        return;
    }

    newDialog.open = true;
}

addEventListener("hashchange", () => openDialogFromHash());
addEventListener("load", function () {
    openDialogFromHash();
});
