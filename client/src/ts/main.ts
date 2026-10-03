import "./theme.ts";

import htmx from "htmx.org";
import type HTMX from "htmx.org";
import "./htmx-extension.ts";
import { domLoaded, initAnchorHeadersFor } from "./lib.ts";

declare namespace globalThis {
	let htmx: typeof HTMX.default;
}
globalThis.htmx = htmx as unknown as typeof HTMX.default;

(htmx as unknown as typeof htmx.default).config
	.methodsThatUseUrlParams.length = 0;

import("htmx-ext-debug");

domLoaded.then(() => initAnchorHeadersFor(document.body));

function formatDates(target: ParentNode): void {
	const dates = target instanceof HTMLTimeElement &&
			target.matches("[data-format-date]")
		? [target]
		: Array.from(
			target.querySelectorAll<HTMLTimeElement>("time[data-format-date]"),
		);
	for (const time of dates) {
		const date = new Date(time.dateTime);
		if (Number.isNaN(date.getTime())) {
			console.error("Invalid date value: %o", time);
			continue;
		}
		time.textContent = new Intl.DateTimeFormat(undefined, {
			dateStyle: "long",
			timeStyle: "short",
		}).format(date);
	}
}

domLoaded.then(() => {
	formatDates(document);
	document.body.addEventListener("htmx:afterSwap", (event) => {
		if (event.target instanceof Element) {
			formatDates(event.target);
		}
	});

	document.body.addEventListener("click", async (event) => {
		const target = event.target;
		if (!(target instanceof Element)) return;

		for (
			const menu of document.querySelectorAll<HTMLElement>(
				".dropdown-content, .app-context-menu",
			)
		) {
			const dropdown = menu.closest<HTMLDetailsElement>("details[open]");
			if (dropdown && !dropdown.contains(target)) {
				dropdown.open = false;
			}
		}

		const dismiss = target.closest<HTMLButtonElement>("[data-close-alert]");
		if (dismiss) {
			dismiss.closest('[role="alert"]')?.remove();
			return;
		}

		const passwordToggle = target.closest<HTMLButtonElement>(
			"[data-toggle-password]",
		);
		if (passwordToggle) {
			const input = passwordToggle.parentElement?.querySelector<
				HTMLInputElement
			>("input");
			if (!input) {
				console.error("Password toggle has no input: %o", passwordToggle);
				return;
			}
			const visibleIcon = passwordToggle.querySelector<HTMLElement>(
				"[data-password-visible]",
			);
			const hiddenIcon = passwordToggle.querySelector<HTMLElement>(
				"[data-password-hidden]",
			);
			if (!visibleIcon || !hiddenIcon) {
				console.error("Password toggle icons are missing: %o", passwordToggle);
				return;
			}
			const isVisible = input.type === "password";
			input.type = isVisible ? "text" : "password";
			visibleIcon.hidden = isVisible;
			hiddenIcon.hidden = !isVisible;
			passwordToggle.setAttribute(
				"aria-label",
				isVisible ? "Hide password" : "Show password",
			);
			passwordToggle.setAttribute(
				"aria-pressed",
				String(isVisible),
			);
			return;
		}

		const copyButton = target.closest<HTMLButtonElement>("[data-copy]");
		if (copyButton) {
			const value = copyButton.dataset.copy;
			if (value === undefined) {
				console.error("Copy button has no value: %o", copyButton);
				return;
			}
			try {
				await navigator.clipboard.writeText(value);
				const label = copyButton.querySelector<HTMLElement>(
					"[data-copy-label]",
				);
				if (label) label.textContent = "Copied";
			} catch (error) {
				console.error("Could not copy invite link: %o", error);
				const label = copyButton.querySelector<HTMLElement>(
					"[data-copy-label]",
				);
				if (label) label.textContent = "Copy failed";
			}
		}
	});
});
