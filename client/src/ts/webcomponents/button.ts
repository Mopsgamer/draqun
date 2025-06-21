import { html, LitElement, type PropertyDeclarations } from "lit";
import { customElement } from "lit/decorators.js";

@customElement("dq-button")
export class DraqunButton extends LitElement {
    protected override createRenderRoot(): HTMLElement | DocumentFragment {
        return this;
    }

    static override properties: PropertyDeclarations = {
        href: { type: String, reflect: true },
    };

    href?: string = undefined;

    #prefix = "";
    #suffix = "";

    override render() {
        const btn = html`
            <button>
                <slot name="prefix">${this.#prefix}</slot>
                ${this.children}
                <slot name="suffix">${this.#suffix}</slot>
            </button>
        `;

        if (this.href) {
            return html`
                <a href="${this.href}">${btn}</a>
            `;
        }

        return btn;
    }
}
