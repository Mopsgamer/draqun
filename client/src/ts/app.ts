import "./main.ts";

import "./chat-mutate.ts";
import "./chat-send.ts";
import "./chat-mutate.ts";
import { domLoaded } from "./lib.ts";

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
	const membersToggler = document.getElementById("members-toggler")!;
	const searchToggler = document.getElementById("search-toggler")!;
	const searchInput = document.getElementById("search-input")!;

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

	if (searchView && searchToggler) {
		const toggleSearch = () => {
			closeAllBut(searchView, secondaryViewList);
			if (searchView.classList.contains("open")) {
				searchInput.classList.remove("collapse");
				(searchInput as any).focus();
			} else {
				searchInput.classList.add("collapse");
			}
		};

		searchToggler.addEventListener("click", toggleSearch);

		document.addEventListener("keydown", (e) => {
			if (
				e.key === "/" &&
				document.activeElement?.tagName !== "INPUT" &&
				document.activeElement?.tagName !== "TEXTAREA" &&
				!(document.activeElement as any)?.tagName?.startsWith("SL-")
			) {
				e.preventDefault();
				if (!searchView.classList.contains("open")) {
					toggleSearch();
				} else {
					(searchInput as any).focus();
				}
			}

			if (
				e.key === "?" &&
				document.activeElement?.tagName !== "INPUT" &&
				document.activeElement?.tagName !== "TEXTAREA" &&
				!(document.activeElement as any)?.tagName?.startsWith("SL-")
			) {
				e.preventDefault();
				(document.getElementById("search-help-dialog") as any)?.show();
			}
		});
	}
});
