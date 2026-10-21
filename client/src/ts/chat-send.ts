import { domLoaded, findLastMessage } from "./lib.ts";

domLoaded.then(() => {
	const form = document.getElementById(
		"send-message-form",
	) as HTMLFormElement | null;

	if (!form) return;

	const input = form.querySelector<HTMLTextAreaElement>(
		"textarea[data-autoresize]",
	);
	const fileInput = document.getElementById(
		"file-attachment-input",
	) as HTMLInputElement | null;
	const previewContainer = document.getElementById("attachment-preview");
	const previewName = document.getElementById("attachment-preview-name");
	const removeBtn = document.getElementById("attachment-preview-remove");

	const resizeInput = () => {
		if (!input) return;
		input.style.height = "auto";
		input.style.height = Math.min(input.scrollHeight, 80) + "px";
	};
	if (input) {
		input.addEventListener("input", resizeInput);
		resizeInput();
	}

	const updatePreview = () => {
		if (fileInput && fileInput.files && fileInput.files.length > 0) {
			const file = fileInput.files[0];
			if (previewName) previewName.textContent = file.name;
			if (previewContainer) previewContainer.classList.remove("hidden");
		} else {
			if (previewContainer) previewContainer.classList.add("hidden");
			if (previewName) previewName.textContent = "";
		}
	};

	if (fileInput) {
		fileInput.addEventListener("change", updatePreview);
	}

	if (removeBtn) {
		removeBtn.addEventListener("click", () => {
			if (fileInput) fileInput.value = "";
			updatePreview();
		});
	}

	form.addEventListener("htmx:afterRequest", () => {
		form.reset();
		updatePreview();
		requestAnimationFrame(resizeInput);
	});

	findLastMessage()?.scrollIntoView();
});
