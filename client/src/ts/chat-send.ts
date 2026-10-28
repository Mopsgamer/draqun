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
		if (!previewContainer) return;
		previewContainer.innerHTML = "";

		if (fileInput && fileInput.files && fileInput.files.length > 0) {
			previewContainer.classList.remove("hidden");
			previewContainer.classList.add("flex");

			Array.from(fileInput.files).forEach((file, idx) => {
				const chip = document.createElement("span");
				chip.className =
					"inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md border border-base-content/15 text-xs font-medium text-base-content/90 bg-surface-card";

				const icon = document.createElement("span");
				icon.className = "icon icon-[lucide--paperclip] size-3.5 text-primary";

				const nameSpan = document.createElement("span");
				nameSpan.className = "truncate max-w-xs";
				nameSpan.textContent = file.name;

				const removeBtn = document.createElement("button");
				removeBtn.type = "button";
				removeBtn.className =
					"btn btn-ghost btn-xs btn-circle p-0 size-4 min-h-0 text-base-content/50 hover:text-error ml-1 cursor-pointer";
				removeBtn.title = "Remove attachment";
				removeBtn.textContent = "✕";

				removeBtn.addEventListener("click", () => {
					const dt = new DataTransfer();
					if (fileInput.files) {
						Array.from(fileInput.files).forEach((f, i) => {
							if (i !== idx) dt.items.add(f);
						});
					}
					fileInput.files = dt.files;
					updatePreview();
				});

				chip.appendChild(icon);
				chip.appendChild(nameSpan);
				chip.appendChild(removeBtn);

				previewContainer.appendChild(chip);
			});
		} else {
			previewContainer.classList.add("hidden");
			previewContainer.classList.remove("flex");
		}
	};

	if (fileInput) {
		fileInput.addEventListener("change", updatePreview);
	}

	const handlePaste = (e: ClipboardEvent) => {
		if (!e.clipboardData || !fileInput) return;

		const items = e.clipboardData.items;
		if (!items) return;

		const pastedFiles: File[] = [];
		for (let i = 0; i < items.length; i++) {
			const item = items[i];
			if (item.kind === "file") {
				const blob = item.getAsFile();
				if (blob) {
					let fileName = blob.name;
					if (!fileName || fileName === "image.png") {
						const ext = blob.type.split("/")[1] || "png";
						fileName = `pasted-image-${Date.now()}-${i}.${ext}`;
					}
					const file = new File([blob], fileName, { type: blob.type });
					pastedFiles.push(file);
				}
			}
		}

		if (pastedFiles.length > 0) {
			const dt = new DataTransfer();
			if (fileInput.files) {
				Array.from(fileInput.files).forEach((f) => dt.items.add(f));
			}
			pastedFiles.forEach((f) => dt.items.add(f));
			fileInput.files = dt.files;
			updatePreview();
		}
	};

	if (input) {
		input.addEventListener("paste", handlePaste);
	}

	form.addEventListener("htmx:afterRequest", () => {
		form.reset();
		updatePreview();
		requestAnimationFrame(resizeInput);
	});

	findLastMessage()?.scrollIntoView();
});
