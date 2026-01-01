import {
  LitElement,
  css,
  html,
} from "https://cdn.jsdelivr.net/gh/lit/dist@3/core/lit-core.min.js";

export class SearchBarElement extends LitElement {
  static properties = {
    action: { type: String },
    method: { type: String },
    api: { type: String },
    placeholder: { type: String },
    name: { type: String },
    minChars: { type: Number },
    debounceMs: { type: Number },

    // internal state
    open: { type: Boolean, state: true },
    loading: { type: Boolean, state: true },
    query: { type: String, state: true },
    data: { type: Object, state: true },
  };

  constructor() {
    super();
    this.action = "/search";
    this.method = "get";
    this.api = "/api/search_bar";
    this.placeholder = "Search";
    this.name = "q";
    this.minChars = 1;
    this.debounceMs = 150;

    this.open = false;
    this.loading = false;
    this.query = "";
    this.data = null;

    this._cache = new Map();
    this._abort = null;
    this._debounceTimer = null;

    this._onDocumentPointerDown = this._onDocumentPointerDown.bind(this);
  }

  // Light DOM: Bootstrap can style dropdown/menu classes without copying CSS into shadow root.
  createRenderRoot() {
    return this;
  }

  connectedCallback() {
    super.connectedCallback();
    document.addEventListener("pointerdown", this._onDocumentPointerDown, true);
  }

  disconnectedCallback() {
    document.removeEventListener(
      "pointerdown",
      this._onDocumentPointerDown,
      true
    );
    this._abort?.abort();
    clearTimeout(this._debounceTimer);
    super.disconnectedCallback();
  }

  _onDocumentPointerDown(e) {
    if (!this.contains(e.target)) this.open = false;
  }

  _onFocusIn() {
    this.open = true;
    // If there is already text, populate results immediately (parity with your focusin behavior).
    const input = this.renderRoot.querySelector('input[type="search"]');
    const q = input?.value?.trim() ?? "";
    if (q.length >= this.minChars) this._scheduleFetch(q);
  }

  _onInput(e) {
    const q = e.target.value.trim();
    this.query = q;

    if (q.length < this.minChars) {
      this.loading = false;
      this.data = null;
      return;
    }

    this.open = true;
    this._scheduleFetch(q);
  }

  _scheduleFetch(q) {
    clearTimeout(this._debounceTimer);
    this._debounceTimer = setTimeout(() => this._fetch(q), this.debounceMs);
  }

  async _fetch(q) {
    // cache hit
    if (this._cache.has(q)) {
      this.data = this._cache.get(q);
      this.loading = false;
      return;
    }

    // abort any in-flight request
    this._abort?.abort();
    this._abort = new AbortController();

    this.loading = true;
    this.data = null;

    const url = new URL(this.api, window.location.origin);
    url.searchParams.set("input", q);

    try {
      const res = await fetch(url.toString(), { signal: this._abort.signal });
      if (!res.ok) throw new Error(`search api failed: ${res.status}`);

      const data = await res.json();
      this._cache.set(q, data);

      // If user typed more while this was in flight, ignore stale responses.
      if (this.query !== q) return;

      this.data = data;
    } catch (err) {
      if (err?.name === "AbortError") return;
      // Optional: surface an error UI
      this.data = { tags: [], items: [], brands: [], _error: true };
    } finally {
      if (this.query === q) this.loading = false;
    }
  }

  _renderSection(title, items, hrefPrefix) {
    if (!items || items.length === 0) return null;

    return html`
      <div class="mb-2">
        <h6 class="dropdown-header">${title}</h6>
        <ul class="list-unstyled mb-0">
          ${items.map(
            (v) => html`
              <li>
                <a
                  class="dropdown-item"
                  href="${hrefPrefix}${encodeURIComponent(v)}"
                >
                  ${v}
                </a>
              </li>
            `
          )}
        </ul>
      </div>
    `;
  }

  render() {
    return html`
      <form
        class="topbar-search d-none d-lg-flex position-relative"
        role="search"
        action="${this.action}"
        method="${this.method}"
        @focusin=${this._onFocusIn}
      >
        <div class="input-group nav-search">
          <span class="input-group-text">
            <svg
              class="icon"
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 640 640"
              aria-hidden="true"
              focusable="false"
            >
              <path
                d="M480 272C480 317.9 465.1 360.3 440 394.7L566.6 521.4C579.1 533.9 579.1 554.2 566.6 566.7C554.1 579.2 533.8 579.2 521.3 566.7L394.7 440C360.3 465.1 317.9 480 272 480C157.1 480 64 386.9 64 272C64 157.1 157.1 64 272 64C386.9 64 480 157.1 480 272zM272 416C351.5 416 416 351.5 416 272C416 192.5 351.5 128 272 128C192.5 128 128 192.5 128 272C128 351.5 192.5 416 272 416z"
              />
            </svg>
            <span class="visually-hidden">Search</span>
          </span>

          <input
            name="${this.name}"
            type="search"
            class="form-control"
            placeholder="${this.placeholder}"
            autocomplete="off"
            @input=${this._onInput}
            @keydown=${(e) => {
              // Optional: Esc closes dropdown
              if (e.key === "Escape") this.open = false;
            }}
          />
        </div>

        ${this.open
          ? html`
              <div
                class="dropdown-menu show dropdown-search w-100 mt-1"
                style="position:absolute; left:0; top:100%; z-index:2000;"
              >
                ${this.loading
                  ? html`<div class="px-3 py-2 text-muted small">
                      Searching…
                    </div>`
                  : this.data
                  ? html`
                      ${this._renderSection("Tags", this.data.tags, "/tag/")}
                      ${this._renderSection("Items", this.data.items, "/item/")}
                      ${this._renderSection(
                        "Brands",
                        this.data.brands,
                        "/brand/"
                      )}
                      ${!this.data.tags?.length &&
                      !this.data.items?.length &&
                      !this.data.brands?.length
                        ? html`<div class="px-3 py-2 text-muted small">
                            No results
                          </div>`
                        : null}
                    `
                  : html`<div class="px-3 py-2 text-muted small">
                      Type to search
                    </div>`}
              </div>
            `
          : null}
      </form>
    `;
  }
}

customElements.define("search-bar", SearchBarElement);

export class AddToClosetButton extends LitElement {
  static properties = {
    brand: { type: String },
    item: { type: String },

    open: { type: Boolean, state: true },
    loading: { type: Boolean, state: true },
    query: { type: String, state: true },
    data: { type: Object, state: true },
  };

  constructor() {
    super();
    this.brand = "";
    this.item = "";
    this.closets = [];

    this._onDocumentPointerDown = this._onDocumentPointerDown.bind(this);
  }

  async _onClick() {
    if (this.open) {
      this.open = false;
      return;
    }

    let req = await fetch("/api/user/closets");
    let closets = await req.json();
    this.closets = closets;
    this.open = true;
    console.log(this.closets);
  }

  /**
   * 
   * @param {PointerEvent} e 
   * @returns 
   */
  async _addToCloset(e) {
    const closetName = e.target.textContent.trim();
    console.log(`Adding ${this.item} by ${this.brand} to closet ${closetName}`);

    let req = await fetch("/api/user/closets/add_item", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        closet_name: closetName,
        item: this.item,
        brand: this.brand,
      }),
    });
    if (!req.ok) {
      console.error("Failed to add item to closet");
      return;
    }
    this.open = false;

    console.log(`Adding ${this.item} by ${this.brand} to closet ${closetName}`);
  }

  connectedCallback() {
    super.connectedCallback();
    document.addEventListener("pointerdown", this._onDocumentPointerDown, true);
  }

  disconnectedCallback() {
    document.removeEventListener(
      "pointerdown",
      this._onDocumentPointerDown,
      true
    );
    this._abort?.abort();
    clearTimeout(this._debounceTimer);
    super.disconnectedCallback();
  }

  _onDocumentPointerDown(e) {
    if (!this.contains(e.target)) this.open = false;
  }

  createRenderRoot() {
    return this;
  }

  render() {
    const btnSelector = ".add-to-closet-trigger";
    let menuStyle = "position:fixed; right:1rem; top:3rem; z-index:2000;"; // fallback
    if (this.open) {
      const btn = this.renderRoot.querySelector(btnSelector);
      if (btn) {
        const r = btn.getBoundingClientRect();
        menuStyle = `position:fixed; left:${Math.max(
          0,
          r.left
        )}px; top:${Math.max(0, r.bottom)}px; min-width:${
          r.width
        }px; z-index:2000;`;
      }
    }

    return html` <button
      @click=${this._onClick}
      type="button"
      class="add-to-closet-trigger btn btn-link position-absolute top-0 end-0 m-2 z-3 p-2 d-flex align-items-center justify-content-center text-light"
      aria-label="Add to bag"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="1.25rem"
        height="1.25rem"
        fill="currentColor"
        class="bi bi-heart"
        viewBox="0 0 16 16"
      >
        <path
          d="m8 2.748-.717-.737C5.6.281 2.514.878 1.4 3.053c-.523 1.023-.641 2.5.314 4.385.92 1.815 2.834 3.989 6.286 6.357 3.452-2.368 5.365-4.542 6.286-6.357.955-1.886.838-3.362.314-4.385C13.486.878 10.4.28 8.717 2.01zM8 15C-7.333 4.868 3.279-3.04 7.824 1.143q.09.083.176.171a3 3 0 0 1 .176-.17C12.72-3.042 23.333 4.867 8 15"
        />
      </svg>
      ${this.open
        ? html`<div
            class="dropdown-menu show dropdown-search mt-1"
            style="${menuStyle}"
          >
            <h3 class="dropdown-header">Select Closet to Add</h3>
            ${this.closets.map(
              (closet) =>
                html`<button @click=${this._addToCloset} class="dropdown-item">
                  ${closet.name}
                </button>`
            )}
          </div>`
        : null}
    </button>`;
  }
}

customElements.define("add-to-closet-button", AddToClosetButton);
