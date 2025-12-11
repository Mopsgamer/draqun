import { domLoaded } from "./lib.ts";
import { renderMarkdown } from "./markdown.ts";

/**
 * Global attribute to mark elements that have been processed for Markdown.
 */
const PROCESSED_ATTR = "data-md-processed";

/**
 * Formats a single message text element, converting Markdown to HTML.
 */
function processMessageText(textElement: Element): void {
    // Skip if already processed
    if (textElement.hasAttribute(PROCESSED_ATTR)) {
        return;
    }

    // Get raw text content
    const rawText = textElement.textContent || "";
    if (!rawText.trim()) {
        textElement.setAttribute(PROCESSED_ATTR, "true");
        return;
    }

    // Render Markdown and set as HTML
    const renderedHtml = renderMarkdown(rawText);
    textElement.innerHTML = renderedHtml;
    textElement.setAttribute(PROCESSED_ATTR, "true");
}

/**
 * Formats all message text elements in a container.
 */
function processAllMessages(container: Element): void {
    const textElements = container.querySelectorAll(".message .text");
    for (const textElement of textElements) {
        processMessageText(textElement);
    }
}

/**
 * Markdown renderer for the chat.
 */
domLoaded.then(() => {
    const chat = document.getElementById("chat");
    if (!chat) return;

    // Process existing messages
    processAllMessages(chat);

    // Observe for new messages (pagination, websocket updates)
    const observer = new MutationObserver((mutations) => {
        for (const mutation of mutations) {
            for (const node of mutation.addedNodes) {
                if (!(node instanceof Element)) continue;

                // only process if the added node is a single message
                if (node.classList.contains("message")) {
                    const textElement = node.querySelector(".text");
                    if (textElement) {
                        processMessageText(textElement);
                    }
                }

                // Check if the added node contains multiple messages & format them
                processAllMessages(node);
            }
        }
    });

    observer.observe(chat, {
        childList: true,
        subtree: true,
    });
});
