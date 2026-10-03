import htmx from "htmx.org";
import { getFormPropData } from "./lib.ts";

const onEvent: htmx.HtmxExtension["onEvent"] = function (name, event): boolean {
	if (name === "htmx:beforeRequest" || name === "htmx:afterRequest") {
		const enable = name === "htmx:beforeRequest";
		let control: EventTarget | null = event.target;
		if (control instanceof HTMLFormElement) {
			const form = control;
			control = form.querySelector<HTMLButtonElement>(
				"button[type=submit]",
			) ??
				Array.from(
					document.querySelectorAll<HTMLButtonElement>(
						"button[type=submit][form]",
					),
				).find((button) => button.form === form) ??
				form;
		}

		if (control instanceof HTMLButtonElement) {
			control.disabled = enable;
			control.classList.toggle("loading", enable);
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
	const form = event.detail.elt;
	if (!(form instanceof HTMLFormElement)) {
		console.groupEnd();
		return true;
	}

	const formData = getFormPropData(form);
	Object.assign(event.detail.formData, formData);
	console.log("configRequest details: %o", event.detail);

	if (!form.checkValidity()) {
		console.error("Form is invalid: %o", form);
		console.groupEnd();
		return false;
	}
	console.groupEnd();
	return true;
};

(htmx as unknown as typeof htmx.default).defineExtension("app", {
	onEvent,
});
