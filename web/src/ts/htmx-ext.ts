import htmx from "htmx.org";
import type { SlRadioGroup, SlRating, SlInput } from "@shoelace-style/shoelace";

function onEvent(name: htmx.HtmxEvent, evt: CustomEvent) { // htmx type definitions sucks
    if (name !== "htmx:configRequest") {
        return
    }

    if (evt.detail.elt.tagName !== "FORM") {
        return
    }

    const form = evt.detail.elt as HTMLFormElement // sucks

    const slElementList = Array.from(
        form.querySelectorAll("sl-radio-group, sl-rating, sl-input")
    ) as Array<(SlRadioGroup | SlRating | SlInput) & HTMLFormElement>

    for (const slElement of slElementList) {
        const isDisabled = !slElement.name || slElement.disabled ||
            slElement.closest("[disabled]")
        if (isDisabled) continue

        const ratingOrInputName = (slElement.tagName === "SL-RATING" || slElement.tagName === "SL-INPUT") &&
            slElement.getAttribute("name");

        let name = slElement.name
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
