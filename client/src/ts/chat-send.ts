import { domLoaded, findLastMessage } from "./lib.ts";

domLoaded.then(() => {
	const form = document.getElementById(
		"send-message-form",
	) as HTMLFormElement | null;

	if (!form) return;

	const input = form.querySelector<HTMLTextAreaElement>(
		"textarea[data-autoresize]",
	);
	const resizeInput = () => {
		if (!input) return;
		input.style.height = "auto";
		input.style.height = Math.min(input.scrollHeight, 80) + "px";
	};
	if (input) {
		input.addEventListener("input", resizeInput);
		resizeInput();
	}

	form.addEventListener(
		"htmx:afterRequest",
		() => {
			form.reset();
			requestAnimationFrame(resizeInput);
		},
	);

	findLastMessage()?.scrollIntoView();
});
