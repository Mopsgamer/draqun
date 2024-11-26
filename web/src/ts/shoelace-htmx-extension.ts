import htmx from "htmx.org";
import { getFormControls, SlButton } from "@shoelace-style/shoelace";

htmx.defineExtension("shoelace", {
    onEvent(name, evt) {
        console.log(name)
        if (name === "htmx:beforeSend"|| name === "htmx:afterRequest") {
            const form = evt.target;
            let button: SlButton | undefined
            if (form instanceof SlButton) {
                button = form
            } else if (form instanceof HTMLFormElement) {
                button = form.querySelector<SlButton>('sl-button[type=submit]') ?? undefined
            }

            if (!button) {
                return;
            }
            button.loading = name === "htmx:beforeSend"
            return;
        }
        if (name !== "htmx:configRequest") {
            return;
        }

        const form = evt.detail.elt;
        if (!(form instanceof HTMLFormElement)) {
            console.groupEnd();
            return;
        }

        for (const slElement of getFormControls(form) as HTMLFormElement[]) {
            const { name, value } = slElement;

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
