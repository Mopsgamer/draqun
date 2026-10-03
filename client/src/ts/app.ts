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
	const membersTogglers = Array.from(
		document.getElementsByClassName("group-members-toggle"),
	).filter((element): element is HTMLButtonElement =>
		element instanceof HTMLButtonElement
	);
	const secondaryViewList = Array.from(
		document.getElementsByClassName("secondary-view"),
	) as HTMLElement[];

	const membersView = document.getElementById("members-view");
	if (membersTogglers.length === 0 || !membersView) return;

	for (const membersToggler of membersTogglers) {
		membersToggler.addEventListener("click", () => {
			closeAllBut(membersView, secondaryViewList);
			const isOpen = membersView.classList.contains("open");
			for (const toggler of membersTogglers) {
				toggler.setAttribute("aria-expanded", String(isOpen));
			}
		});
	}
});
