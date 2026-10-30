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

	// Mention Autocompletion
	const popover = document.getElementById("mention-popover");
	let selectedIndex = -1;
	let currentItems: Array<{
		type: string;
		id: number;
		moniker: string;
		name: string;
		syntax: string;
	}> = [];
	let matchStart = -1;
	let queryLen = 0;

	const hidePopover = () => {
		if (popover) {
			popover.classList.add("hidden");
			popover.innerHTML = "";
		}
		selectedIndex = -1;
		currentItems = [];
	};

	const renderPopover = () => {
		if (!popover) return;
		if (currentItems.length === 0) {
			hidePopover();
			return;
		}

		popover.innerHTML = "";
		popover.classList.remove("hidden");

		currentItems.forEach((item, index) => {
			const row = document.createElement("div");
			row.className = `flex items-center justify-between px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors ${
				index === selectedIndex
					? "bg-primary/15 text-primary font-medium"
					: "hover:bg-base-content/10"
			}`;

			const left = document.createElement("div");
			left.className = "flex items-center gap-2 min-w-0";

			const badge = document.createElement("span");
			let typeLabel = "@user";
			let badgeClass = "bg-primary/10 text-primary";

			switch (item.type) {
				case "u":
					typeLabel = "user";
					badgeClass = "bg-primary/10 text-primary";
					break;
				case "r":
					typeLabel = "role";
					badgeClass = "bg-secondary/10 text-secondary";
					break;
				case "c":
					typeLabel = "channel";
					badgeClass = "bg-accent/10 text-accent";
					break;
				case "g":
					typeLabel = "group";
					badgeClass = "bg-info/10 text-info";
					break;
			}
			badge.className = `px-1.5 py-0.5 rounded text-[10px] font-bold ${badgeClass}`;
			badge.textContent = typeLabel;

			const label = document.createElement("span");
			label.className = "truncate font-medium text-base-content";
			label.textContent = `@${item.moniker}`;

			left.appendChild(badge);
			left.appendChild(label);

			const right = document.createElement("span");
			right.className = "text-[11px] text-base-content/50 truncate ml-2 font-mono";
			right.textContent = item.syntax;

			row.appendChild(left);
			row.appendChild(right);

			row.addEventListener("mousedown", (e) => {
				e.preventDefault();
				selectItem(index);
			});

			popover.appendChild(row);
		});
	};

	const selectItem = (index: number) => {
		if (!input || index < 0 || index >= currentItems.length) return;
		const item = currentItems[index];

		const val = input.value;
		const before = val.slice(0, matchStart);
		const after = val.slice(matchStart + queryLen);

		const inserted = `${item.syntax} `;
		input.value = before + inserted + after;

		const newCursor = before.length + inserted.length;
		input.setSelectionRange(newCursor, newCursor);

		hidePopover();
		input.focus();
		resizeInput();
	};

	const checkMentionTrigger = async () => {
		if (!input) return;
		const cursorPos = input.selectionStart;
		const val = input.value;
		const textBefore = val.slice(0, cursorPos);

		const match = textBefore.match(/(?:^|\s)@([a-zA-Z0-9._-]*)$/);
		if (!match) {
			hidePopover();
			return;
		}

		const fullMatch = match[0];
		const q = match[1];

		matchStart = cursorPos - fullMatch.length + (fullMatch.startsWith(" ") || fullMatch.startsWith("\n") ? 1 : 0);
		queryLen = cursorPos - matchStart;

		let groupId = input.dataset.groupId;
		let dmChannelId = input.dataset.dmChannelId;
		const action = form.getAttribute("hx-post") || "";

		if (!groupId && !dmChannelId) {
			const gMatch = action.match(/\/groups\/(\d+)/);
			if (gMatch) groupId = gMatch[1];

			const dmMatch = action.match(/\/groups\/dm\/channels\/(\d+)/);
			if (dmMatch) dmChannelId = dmMatch[1];
		}

		let fetchUrl = "";
		if (groupId) {
			fetchUrl = `/groups/${groupId}/mention-autocomplete?q=${encodeURIComponent(q)}`;
		} else if (dmChannelId) {
			fetchUrl = `/groups/dm/channels/${dmChannelId}/mention-autocomplete?q=${encodeURIComponent(q)}`;
		} else {
			hidePopover();
			return;
		}

		try {
			const res = await fetch(fetchUrl);
			if (!res.ok) {
				hidePopover();
				return;
			}
			const data = await res.json();
			currentItems = Array.isArray(data) ? data : [];
			selectedIndex = currentItems.length > 0 ? 0 : -1;
			renderPopover();
		} catch {
			hidePopover();
		}
	};

	if (input) {
		input.addEventListener("keyup", (e) => {
			if (["ArrowUp", "ArrowDown", "Enter", "Tab", "Escape"].includes(e.key) && currentItems.length > 0) {
				return;
			}
			checkMentionTrigger();
		});

		input.addEventListener("keydown", (e) => {
			if (currentItems.length === 0 || !popover || popover.classList.contains("hidden")) {
				return;
			}

			if (e.key === "ArrowDown") {
				e.preventDefault();
				selectedIndex = (selectedIndex + 1) % currentItems.length;
				renderPopover();
			} else if (e.key === "ArrowUp") {
				e.preventDefault();
				selectedIndex = (selectedIndex - 1 + currentItems.length) % currentItems.length;
				renderPopover();
			} else if (e.key === "Enter" || e.key === "Tab") {
				if (selectedIndex >= 0 && selectedIndex < currentItems.length) {
					e.preventDefault();
					selectItem(selectedIndex);
				}
			} else if (e.key === "Escape") {
				e.preventDefault();
				hidePopover();
			}
		});

		input.addEventListener("blur", () => {
			setTimeout(hidePopover, 150);
		});
	}

	form.addEventListener("htmx:afterRequest", () => {
		form.reset();
		updatePreview();
		hidePopover();
		requestAnimationFrame(resizeInput);
	});

	findLastMessage()?.scrollIntoView();
});
