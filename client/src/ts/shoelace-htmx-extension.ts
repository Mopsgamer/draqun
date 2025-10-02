import htmx from "htmx.org";
import type HTMX from "htmx.org";
import { getFormPropData } from "./lib.ts";
import { SlButton, SlMenuItem } from "@shoelace-style/shoelace";

const onEvent: HTMX.HtmxExtension["onEvent"] = function (name, event): boolean {
    if (name === "htmx:beforeRequest" || name === "htmx:afterRequest") {
        const enable = name === "htmx:beforeRequest";

        let form = event.target || {} as object;
        if (event.target instanceof HTMLFormElement) {
            form = (event.target.querySelector('sl-button[type=submit]') || document.querySelector('sl-button[form='+event.target.id+'][type=submit]') || {}) as object
        }

        if (form instanceof SlButton || form instanceof SlMenuItem) {
            form.loading = enable
            form.disabled = enable
        }
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

(htmx as unknown as typeof HTMX.default).defineExtension("shoelace", {
    onEvent,
});
