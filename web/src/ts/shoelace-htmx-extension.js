import * as HTMX from "htmx.org";
// @deno-types="npm:@types/diff"
// import * as diff from "diff";
import { SlButton } from "@shoelace-style/shoelace";
import { getFormPropData } from "./lib.ts";

HTMX.defineExtension("shoelace", {
    onEvent(
        name,
        event,
    ) {
        if (name === "htmx:beforeRequest" || name === "htmx:afterRequest") {
            const form = event.target;
            /** @type {SlButton | null | undefined} */
            let button;
            if (form instanceof SlButton) {
                button = form;
            } else if (form instanceof HTMLFormElement) {
                button = document.querySelector(`sl-button[form=${form.id}][type=submit]`);
                button ??= form.querySelector(`sl-button[type=submit]`);
            }

            if (!button) {
                return true;
            }

            const enable = name === "htmx:beforeRequest"
            button.loading = enable;
            button.disabled = enable;
            return true;
        }
        if (name !== "htmx:configRequest") {
            return true;
        }

        if (!(event instanceof CustomEvent)) {
            console.groupEnd();
            return true;
        }
        const { detail } = event;
        const form = detail.elt;
        if (!(form instanceof HTMLFormElement)) {
            console.groupEnd();
            return true;
        }

        Object.assign(detail.parameters, getFormPropData(form));

        // Prevent form submission if one or more fields are invalid.
        // form is always a form as per the main if statement
        if (!form.checkValidity()) {
            console.error("Form is invalid: %o", form);
            console.groupEnd();
            return false;
        }
        console.groupEnd();
        return true;
    },
});
