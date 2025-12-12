import DOMPurify from "dompurify";
import { Marked } from "marked";

/**
 * Configured Marked instance with safe defaults.
 * HTML passthrough is disabled to prevent XSS.
 */
const marked = new Marked({
    async: false,
    gfm: true, // GitHub Flavored Markdown
    breaks: true, // Convert newlines to <br>
});

/**
 * Sanitization configuration for DOMPurify.
 * Allows safe HTML elements typically used in Markdown.
 */
const PURIFY_CONFIG = {
    ALLOWED_TAGS: [
        "p",
        "br",
        "strong",
        "b",
        "em",
        "i",
        "u",
        "s",
        "del",
        "code",
        "pre",
        "blockquote",
        "ul",
        "ol",
        "li",
        "a",
        "h1",
        "h2",
        "h3",
        "h4",
        "h5",
        "h6",
        "hr",
    ],
    ALLOWED_ATTR: ["href", "title", "target", "rel"],
    ALLOW_DATA_ATTR: false,
    ADD_ATTR: ["target", "rel"],
    RETURN_TRUSTED_TYPE: false,
};

/**
 * DOMPurify hook to make links safe (always open in new tab & prevent opener access).
 */
DOMPurify.addHook("afterSanitizeAttributes", (node: Element) => {
    if (node.tagName === "A") {
        node.setAttribute("target", "_blank");
        node.setAttribute("rel", "noopener noreferrer");
    }
});

/**
 * Render Markdown text to sanitized HTML.
 * @param text - Raw Markdown text
 * @returns Sanitized HTML string
 */
export function renderMarkdown(text: string): string {
    // Markdown -> HTML
    const rawHtml = marked.parse(text) as string;

    // HTML -> Sanitized HTML
    const sanitizedHtml = DOMPurify.sanitize(rawHtml, PURIFY_CONFIG) as string;

    return sanitizedHtml;
}
