import "./main.ts";

import "./chat-mutate.ts";
import "./chat-send.ts";
import { domLoaded } from "./lib.ts";
import type { SlDialog, SlInput } from "@shoelace-style/shoelace";

import("htmx-ext-ws");

function closeAllBut(
	element: HTMLElement,
	secondaryViewList: HTMLElement[],
): void {
	element.classList.toggle("open");
	for (const secondary of secondaryViewList) {
		if (secondary !== element) {
			secondary.classList.remove("open");
		}
	}
}

domLoaded.then(() => {
	const membersToggler = document.getElementById("members-toggler");
	const searchToggler = document.getElementById("search-toggler");
	const searchInput = document.getElementById("search-input") as SlInput | null;

	const secondaryViewList = Array.from(
		document.getElementsByClassName("secondary-view"),
	) as HTMLElement[];

	const membersView = document.getElementById("members-view");
	const searchView = document.getElementById("search-view");

	if (membersView && membersToggler) {
		membersToggler.addEventListener("click", () => {
			closeAllBut(membersView, secondaryViewList);
		});
	}

	if (searchView && searchToggler && searchInput) {
		const toggleSearch = () => {
			closeAllBut(searchView, secondaryViewList);
			if (searchView.classList.contains("open")) {
				searchInput.classList.remove("collapse");
				searchInput.focus();
			} else {
				searchInput.classList.add("collapse");
			}
		};

		searchToggler.addEventListener("click", toggleSearch);

		document.addEventListener("keydown", (e) => {
			const activeElement = document.activeElement;
			const isTyping = activeElement?.tagName === "INPUT" ||
				activeElement?.tagName === "TEXTAREA" ||
				!!activeElement?.tagName?.startsWith("SL-");

			if (e.key === "/" && !isTyping) {
				e.preventDefault();
				if (!searchView.classList.contains("open")) {
					toggleSearch();
				} else {
					searchInput.focus();
				}
			}

			if (e.key === "?" && !isTyping) {
				e.preventDefault();
				const dialog = document.getElementById("search-help-dialog") as
					| SlDialog
					| null;
				if (dialog) {
					dialog.show();
				}
			}
		});
	}
});
