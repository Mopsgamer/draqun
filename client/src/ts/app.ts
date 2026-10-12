import "./main.ts";

import "./chat-mutate.ts";
import "./chat-send.ts";
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
	const channelsTogglers = Array.from(
		document.getElementsByClassName("group-channels-toggle"),
	).filter((element): element is HTMLButtonElement =>
		element instanceof HTMLButtonElement
	);
	const secondaryViewList = Array.from(
		document.getElementsByClassName("secondary-view"),
	) as HTMLElement[];

	const membersView = document.getElementById("members-view");
	const channelsView = document.getElementById("channels-view");

	if (membersView && membersTogglers.length > 0) {
		for (const membersToggler of membersTogglers) {
			membersToggler.addEventListener("click", () => {
				if (channelsView) channelsView.classList.remove("open");
				closeAllBut(membersView, secondaryViewList);
				const isOpen = membersView.classList.contains("open");
				for (const toggler of membersTogglers) {
					toggler.setAttribute("aria-expanded", String(isOpen));
				}
			});
		}
	}

	if (channelsView && channelsTogglers.length > 0) {
		for (const channelsToggler of channelsTogglers) {
			channelsToggler.addEventListener("click", () => {
				if (membersView) membersView.classList.remove("open");
				channelsView.classList.toggle("open");
				channelsView.classList.toggle("closed");
				const isOpen = channelsView.classList.contains("open");
				for (const toggler of channelsTogglers) {
					toggler.setAttribute("aria-expanded", String(isOpen));
				}
			});
		}
	}
});
