import htmx from "htmx.org";
import type { SlInput, SlRadioGroup, SlRating } from "@shoelace-style/shoelace";

htmx.defineExtension("shoelace", {
    onEvent(name, evt) {
        if (name === "htmx:configRequest") {
            console.group("HTMX event: %s", name);
        } else {
            console.log("HTMX event: %s", name);
            return;
        }

        if (!(evt.detail.elt instanceof HTMLFormElement)) {
            console.groupEnd();
            return;
        }

        const form = evt.detail.elt as HTMLFormElement;

        const slElementList = Array.from(
            form.querySelectorAll(
                "sl-radio-group, sl-rating, sl-input, sl-select",
            ),
        ) as Array<(SlRadioGroup | SlRating | SlInput) & HTMLFormElement>;

        for (const slElement of slElementList) {
            const isDisabled = slElement.disabled ||
                slElement.closest("[disabled]:not([disabled=false])");
            const { name, value } = slElement;

            if (isDisabled) {
                console.log("Form data skip (disabled): %o", name);
                continue;
            }

            if (!name) {
                console.error(
                    "Form shoelace element does not have the 'name' attribute, but enabled: %o",
                    slElement,
                );
                continue;
            }

            console.log("Form data set: %o %o", name, value);
            evt.detail.parameters[name] = value;
        }
        console.log(
            "Event detail parameters (form data):",
            evt.detail.parameters,
        );

        // Prevent form submission if one or more fields are invalid.
        // form is always a form as per the main if statement
        if (!form.checkValidity()) {
            console.error("Form is invalid: %o", form);
            console.groupEnd();
            return false;
        }
        console.groupEnd();
    },
});
