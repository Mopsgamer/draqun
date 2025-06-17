import htmx from "htmx.org";
import type HTMX from "htmx.org";
import { SlButton } from "@shoelace-style/shoelace";
import { getFormPropData } from "./lib.ts";

const onEvent: HTMX.HtmxExtension["onEvent"] = function (name, event) {
    if (name === "htmx:beforeRequest" || name === "htmx:afterRequest") {
        const form = event.target;
        let button: SlButton | undefined;
        if (form instanceof SlButton) {
            button = form;
        } else if (form instanceof HTMLFormElement) {
            button = document.querySelector<SlButton>(
                `sl-button[form=${form.id}][type=submit]`,
            ) ?? undefined;
            button ??= form.querySelector<SlButton>(`sl-button[type=submit]`) ??
                undefined;
        }

        if (!button) {
            return true;
        }

        const enable = name === "htmx:beforeRequest";
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
};

(htmx as unknown as typeof HTMX.default).defineExtension("shoelace", { onEvent });
