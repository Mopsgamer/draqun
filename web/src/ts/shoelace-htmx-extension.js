import * as HTMX from "htmx.org";
// @deno-types="npm:@types/diff"
// import * as diff from "diff";
import { getFormControls, SlButton } from "@shoelace-style/shoelace";

globalThis.htmx = HTMX;

HTMX.on("htmx:wsConfigSend", (
    event,
) => {
    const form = event.target;

    if (!(form instanceof HTMLFormElement) || !(event instanceof CustomEvent)) {
        return;
    }

    const { detail } = event;

    Object.assign(detail.parameters, getFormPropData(form, true), {
        Type: form.id,
    });
});

// TODO: htmx websocket does not work
// fixes websocket, idk why, it does not replaces the dom
// we are sending only the string content as examples and docs saying
// HTMX.on("htmx:wsAfterMessage", (
//     event,
// ) => {
//     if (!(event instanceof CustomEvent)) {
//         return;
//     }

//     const { detail } =
//         /**@type {CustomEvent<{elt: HTMLElement, message: string}>}*/ (event);
//     const { elt } = detail;
//     let { message } = detail;
//     message = message.replaceAll('\r', '')

//     const chat = document.getElementById("chat");
//     const diffChars = diff.diffChars(elt.innerHTML, message);
//     const isSame = diffChars.some((c) => c.added || c.removed);
//     if (isSame) {
//         return;
//     }

//     console.dir(diffChars);
//     const h = chat.scrollHeight;
//     elt.innerHTML = message;
//     chat.scrollTo(h, 0);
// });

HTMX.defineExtension("shoelace", {
    onEvent(
        name,
        event,
    ) {
        if (name === "htmx:beforeSend" || name === "htmx:afterRequest") {
            const form = event.target;
            let button;
            if (form instanceof SlButton) {
                button = form;
            } else if (form instanceof HTMLFormElement) {
                button = form.querySelector("sl-button[type=submit]") ??
                    undefined;
            }

            if (!button) {
                return true;
            }
            button.loading = name === "htmx:beforeSend";
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

function getFormPropData(
    form,
    capital = false,
) {
    const data = {};
    for (const slElement of getFormControls(form)) {
        let { name } = slElement;
        const { value } = slElement;

        if (!name) {
            continue;
        }

        if (capital) {
            name = name[0].toUpperCase() + name.substring(1);
        }
        data[name] = value;
    }

    return data;
}
