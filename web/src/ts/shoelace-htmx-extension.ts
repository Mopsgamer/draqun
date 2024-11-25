import htmx from "htmx.org";
import { getFormControls } from "@shoelace-style/shoelace";

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
