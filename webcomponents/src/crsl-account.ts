import { html, LitElement, TemplateResult } from "lit";
import { customElement, property, state } from "lit/decorators.js";

@customElement("csrl-account")
export class CsrlAccount extends LitElement {
    @state()
    open = false;

    @state()
    addressesOpen = false;

    @property({ type: String })
    username = "";
    @property({ type: String })
    firstName = "";
    @property({ type: String })
    lastName = "";
    @property({ type: String })
    email = "";

    override createRenderRoot() {
        return this;
    }

    saveProfile(e : Event) {
        e.preventDefault();
        console.log("Save profile clicked");
        this.open = false;
    }

    get editProfile(): TemplateResult {
        return html`
        <form>
            <div class="mb-3">
                <label for="firstName" class="form-label">First Name</label>
                <input type="text" class="form-control" id="firstName" .value=${this.firstName}>
            </div>
            <div class="mb-3">
                <label for="lastName" class="form-label">Last Name</label>
                <input type="text" class="form-control" id="lastName" .value=${this.lastName}>
            </div>
            <button type="submit" class="btn btn-primary" @click=${this.saveProfile}>Save changes</button>
        </form>`;
    }

    get editPassword(): TemplateResult {
        return html`
        <form>
            <div class="mb-3">
                <label for="currentPassword" class="form-label">Current Password</label>
                <input type="password" class="form-control" id="currentPassword">
            </div>
            <div class="mb-3">
                <label for="newPassword" class="form-label">New Password</label>
                <input type="password" class="form-control" id="newPassword">
            </div>
            <div class="mb-3">
                <label for="confirmPassword" class="form-label">Confirm New Password</label>
                <input type="password" class="form-control" id="confirmPassword">
            </div>
            <button type="submit" class="btn btn-primary">Change Password</button>
        </form>`;
    }

    override render() {
        return html`
        <div class="card border-0 shadow-sm rounded-4">
            <div class="card-body p-4">
                <div class="d-flex align-items-center gap-3 mb-3">
                    <div class="rounded-circle bg-light d-inline-flex align-items-center justify-content-center"
                        style="width: 52px; height: 52px;">
                        <span class="fw-semibold text-secondary">
                            ${this.firstName.charAt(0)}${
            this.lastName.charAt(0)
        }
                        </span>
                    </div>
                    <div class="flex-grow-1">
                        <div class="fw-semibold">
                            ${this.firstName} ${this.lastName}
                        </div>
                        <div class="small text-muted">
                            ${this.email}
                        </div>
                    </div>
                </div>

                <hr class="my-4">

                <div class="d-grid gap-2">
                    <button class="btn btn-outline-secondary" @click=${() =>
            this.open = !this.open}>
                        Edit profile
                    </button>
                    <crsl-modal ?open=${this.open}>${this.editProfile}</crsl-modal>

                    <a href="/account/security" class="btn btn-outline-secondary">
                        Login & security
                    </a>
                    <button class="btn btn-outline-secondary" 
                        @click=${() => this.addressesOpen = !this.addressesOpen}
                    >
                        Addresses
                    </button>
                    <crsl-modal ?open=${this.addressesOpen}><div class="btn m-4">Edit Address Here</div></crsl-modal>

                </div>
            </div>
        </div>`;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        "csrl-account": CsrlAccount;
    }
}
