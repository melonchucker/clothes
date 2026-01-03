import { css, html, LitElement } from "lit";
import { customElement, property, query } from "lit/decorators.js";

@customElement("crsl-modal")
export class CrslModal extends LitElement {
    static override styles = css`
    dialog {
      padding: 0;
      border: none;
      border-radius: 8px;
      box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    }
    dialog::backdrop {
      background: rgba(0, 0, 0, 0.5);
    }
    .content {
      position: relative;
      padding: 1rem;
    }
    button.close {
      position: absolute;
      top: 0.5rem;
      right: 0.5rem;
      background: none;
      border: none;
      font-size: 1.5rem;
      cursor: pointer;
    }
  `;

    // reflect allows <csl-modal open> in plain HTML to keep attribute/property in sync
    @property({ type: Boolean, reflect: true })
    open = false;

    @query("dialog")
    private _dialog!: HTMLDialogElement;

    protected override updated(changed: Map<string, unknown>) {
        if (!changed.has("open")) return;
        if (!this._dialog) return;

        if (this.open) {
            // showModal gives proper modal behavior (focus, ESC, backdrop)
            if (!this._dialog.open) this._dialog.showModal();
        } else {
            if (this._dialog.open) this._dialog.close();
        }
    }

    private _onDialogClick(e: MouseEvent) {
        // If the click target is the dialog itself (not inside content), it's a backdrop click
        if (e.target === this._dialog) {
            this.open = false;
        }
    }

    private _onCloseClick() {
        this.open = false;
    }

    private _onDialogClose() {
        // Keep property in sync if user closes via ESC or other native close action
        if (this.open) this.open = false;
    }

    override render() {
        return html`
      <dialog @click=${this._onDialogClick} @close=${this._onDialogClose}>
        <div class="content">
          <button class="close" @click=${this._onCloseClick} aria-label="Close modal">
            &times;
          </button>
          <slot></slot>
        </div>
      </dialog>
    `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        "crsl-modal": CrslModal;
    }
}
