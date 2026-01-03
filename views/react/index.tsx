import React from "react";
import { createRoot } from "react-dom/client";
import useSWR, { SWRConfig } from "swr";

async function fetcher(resource: string, init?: RequestInit) {
    const res = await fetch(resource, init);
    return res.json();
}

type ClosetItem = {
    base_item_name: string;
    brand_name?: string;
    thumbnail_url?: string;
};

type Closet = {
    name: string;
    items: ClosetItem[];
};

type SiteUser = {
    firstName: string;
    lastName: string;
    email: string;
    isAdmin?: boolean;
    isStaff?: boolean;
    isActive?: boolean;
};

type Order = {
    id: string;
    date: string;
    status: string;
    total: string;
    badges?: string[];
};

function Hero() {
    return (
        <section
            className="py-5 py-lg-6 mb-4"
            style={{
                background:
                    "radial-gradient(circle at top left, rgba(255, 214, 165, 0.85), transparent 55%), radial-gradient(circle at top right, rgba(255, 180, 220, 0.8), transparent 55%), linear-gradient(135deg, #fff8f2, #fef6ff)",
            }}
        >
            <div className="container">
                <div className="d-flex flex-column flex-lg-row align-items-start align-items-lg-center justify-content-between gap-3">
                    <div>
                        <p className="small text-uppercase text-secondary mb-1">
                            Your account
                        </p>
                        <h1 className="display-6 fw-bold mb-2">
                            Welcome back!
                        </h1>
                        <p className="text-muted mb-0">
                            Manage your membership, deliveries, and account
                            details.
                        </p>
                    </div>

                    <div className="d-flex flex-column flex-sm-row gap-2">
                        <a
                            href="/browse"
                            className="btn btn-primary btn-lg px-4"
                        >
                            Start styling
                        </a>
                        <a
                            href="/sign-out"
                            className="btn btn-outline-secondary btn-lg px-4"
                        >
                            Sign out
                        </a>
                    </div>
                </div>
            </div>
        </section>
    );
}

function ProfileCard({ user }: { user: SiteUser }) {
    const initials = `${user.firstName.charAt(0)}${user.lastName.charAt(0)}`
        .toUpperCase();
    return (
        <div className="card border-0 shadow-sm rounded-4">
            <div className="card-body p-4">
                <div className="d-flex align-items-center gap-3 mb-3">
                    <div
                        className="rounded-circle bg-light d-inline-flex align-items-center justify-content-center"
                        style={{ width: 52, height: 52 }}
                    >
                        <span className="fw-semibold text-secondary">
                            {initials}
                        </span>
                    </div>
                    <div className="flex-grow-1">
                        <div className="fw-semibold">
                            {user.firstName} {user.lastName}
                        </div>
                        <div className="small text-muted">{user.email}</div>
                    </div>
                </div>

                <div className="d-flex flex-wrap gap-2 mb-3">
                    {user.isAdmin
                        ? (
                            <span className="badge rounded-pill text-bg-danger-subtle text-danger-emphasis px-3 py-2">
                                Admin
                            </span>
                        )
                        : user.isStaff
                        ? (
                            <span className="badge rounded-pill text-bg-warning-subtle text-warning-emphasis px-3 py-2">
                                Staff
                            </span>
                        )
                        : (
                            <span className="badge rounded-pill text-bg-primary-subtle text-primary-emphasis px-3 py-2">
                                Member
                            </span>
                        )}

                    <span className="badge rounded-pill text-bg-success-subtle text-success-emphasis px-3 py-2">
                        {user.isActive ? "Active" : "Inactive"}
                    </span>
                </div>
            </div>
        </div>
    );
}

function ClosetsCard({ closets }: { closets: Closet[] }) {
    const { data: swrClosets, mutate } = useSWR<Closet[]>("/api/user/closets");
    const displayClosets = swrClosets || closets;
    console.log(displayClosets);

    const handleCreateCloset = async (
        event: React.FormEvent<HTMLFormElement>,
    ) => {
        event.preventDefault();
        const formData = new FormData(event.currentTarget);
        const closetName = formData.get("closet_name") as string;

        const res = await fetch("/api/user/closets", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ closet_name: closetName }),
        });

        if (res.ok) {
            // revalidate SWR cache
            await mutate?.();
            // clear the form
            event.currentTarget.reset();
        } else {
            console.error("Failed to create closet");
        }
    };

    const handleDeleteCloset = async (closetName: string) => {
        const res = await fetch("/api/user/closets", {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ closet_name: closetName }),
        });

        if (res.ok) {
            // revalidate SWR cache
            await mutate?.();
        } else {
            console.error("Failed to delete closet");
        }
    };

    return (
        <div className="card border-0 shadow-sm rounded-4 mt-4">
            <div className="card-body p-4">
                <h2 className="h6 fw-semibold mb-1">Closets</h2>
                <p className="small text-muted mb-3">
                    Manage your saved closets
                </p>

                <div className="d-grid gap-2 mb-3">
                    <form
                        className="d-flex flex-column flex-lg-row gap-2"
                        onSubmit={handleCreateCloset}
                    >
                        <input
                            className="form-control flex-grow-1"
                            name="closet_name"
                            type="text"
                            placeholder="New closet name"
                            required
                        />
                        <button
                            type="submit"
                            className="btn btn-outline-secondary w-100 w-lg-auto"
                        >
                            Create new closet
                        </button>
                    </form>
                </div>
                <div className="accordion" id="accordionExample">
                    {displayClosets.map((c) => (
                        <div className="accordion-item" key={c.name}>
                            <h2 className="accordion-header">
                                <button
                                    className="accordion-button collapsed"
                                    type="button"
                                    data-bs-toggle="collapse"
                                    data-bs-target={`#${
                                        c.name.replace(/\s+/g, "-")
                                    }`}
                                    aria-expanded="false"
                                    aria-controls={c.name}
                                >
                                    {c.name} - {c.items.length} items
                                </button>
                            </h2>
                            <div
                                id={c.name.replace(/\s+/g, "-")}
                                className="accordion-collapse collapse"
                                data-bs-parent="#accordionExample"
                            >
                                <div className="accordion-body">
                                    <button
                                        className="btn btn-danger btn-sm mb-3"
                                        onClick={() =>
                                            handleDeleteCloset(c.name)}
                                    >
                                        Delete
                                    </button>
                                    {c.items.map((it, idx) => (
                                        <div
                                            className="mb-2 d-flex align-items-center gap-3"
                                            key={idx}
                                        >
                                            {it.thumbnail_url
                                                ? (
                                                    <img
                                                        src={`/static/images/${it.thumbnail_url}`}
                                                        alt={it.base_item_name}
                                                        className="img-fluid"
                                                        style={{ maxWidth: 64 }}
                                                    />
                                                )
                                                : null}
                                            <div>
                                                <div className="fw-semibold">
                                                    {it.base_item_name}
                                                </div>
                                                {it.brand_name
                                                    ? (
                                                        <div className="small text-muted">
                                                            — {it.brand_name}
                                                        </div>
                                                    )
                                                    : null}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
}

function MembershipCard() {
    return (
        <div className="card border-0 shadow-sm rounded-4 mt-4">
            <div className="card-body p-4">
                <h2 className="h6 fw-semibold mb-1">Membership</h2>
                <p className="small text-muted mb-3">Current plan and perks</p>

                <div className="small">
                    <div className="d-flex justify-content-between">
                        <span className="text-muted">Plan</span>
                        <span className="fw-semibold">Essential (Demo)</span>
                    </div>
                    <div className="d-flex justify-content-between mt-2">
                        <span className="text-muted">Next billing</span>
                        <span className="fw-semibold">April 15, 2025</span>
                    </div>
                    <div className="d-flex justify-content-between mt-2">
                        <span className="text-muted">Shipping</span>
                        <span className="fw-semibold">Same-day eligible</span>
                    </div>
                </div>

                <div className="d-grid gap-2 mt-4">
                    <a href="/pricing" className="btn btn-primary">
                        Manage membership
                    </a>
                    <a href="/help" className="btn btn-outline-secondary">
                        Help & support
                    </a>
                </div>
            </div>
        </div>
    );
}

function CurrentShipmentCard() {
    return (
        <div className="card border-0 shadow-lg rounded-4">
            <div className="card-body p-4 p-lg-5">
                <div className="d-flex flex-column flex-md-row justify-content-between align-items-md-start gap-3">
                    <div>
                        <p className="small text-uppercase text-secondary mb-1">
                            In progress
                        </p>
                        <h2 className="h4 fw-bold mb-1">
                            Your current shipment
                        </h2>
                        <p className="text-muted mb-0">
                            Track delivery and return status.
                        </p>
                    </div>
                    <a
                        href="/account/orders"
                        className="small text-decoration-none"
                    >
                        View all orders →
                    </a>
                </div>

                <hr className="my-4" />

                <div className="row g-3 align-items-center">
                    <div className="col-md-7">
                        <div className="d-flex flex-wrap gap-2 mb-2">
                            <span className="badge text-bg-primary-subtle text-primary-emphasis">
                                Order #FP-123456
                            </span>
                            <span className="badge text-bg-light text-dark">
                                Out for delivery
                            </span>
                            <span className="badge text-bg-warning-subtle text-warning-emphasis">
                                Same-day
                            </span>
                        </div>

                        <div className="small text-muted">
                            <div>
                                <span className="fw-semibold text-dark">
                                    Delivery:
                                </span>{" "}
                                Today · 4–8 PM
                            </div>
                            <div className="mt-1">
                                <span className="fw-semibold text-dark">
                                    Return by:
                                </span>{" "}
                                March 28, 2025
                            </div>
                        </div>
                    </div>

                    <div className="col-md-5">
                        <div className="d-grid gap-2">
                            <a href="#" className="btn btn-primary">
                                Track shipment
                            </a>
                            <a href="#" className="btn btn-outline-secondary">
                                Start a return
                            </a>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

function RecentOrdersCard({ orders }: { orders: Order[] }) {
    return (
        <div className="card border-0 shadow-sm rounded-4 mt-4">
            <div className="card-body p-4">
                <h2 className="h5 fw-semibold mb-3">Recent orders</h2>

                <div className="table-responsive">
                    <table className="table align-middle mb-0">
                        <thead>
                            <tr className="small text-muted">
                                <th>Order</th>
                                <th>Date</th>
                                <th>Status</th>
                                <th className="text-end">Total</th>
                            </tr>
                        </thead>
                        <tbody>
                            {orders.map((o) => (
                                <tr key={o.id}>
                                    <td className="fw-semibold">#{o.id}</td>
                                    <td className="small text-muted">
                                        {o.date}
                                    </td>
                                    <td>
                                        <span className="badge text-bg-light text-dark">
                                            {o.status}
                                        </span>
                                    </td>
                                    <td className="text-end fw-semibold">
                                        {o.total}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}

function App() {
    // fetch data from API endpoints here
    useSWR<SiteUser>("/api/user");
    useSWR<Closet[]>("/api/closets");
    useSWR<Order[]>("/api/orders");

    // Static dummy data based on the provided template
    const user: SiteUser = {
        firstName: "Jordan",
        lastName: "Melnychuk",
        email: "jordan@example.com",
        isAdmin: false,
        isStaff: false,
        isActive: true,
    };

    const closets: Closet[] = [
        {
            name: "Summer Closet",
            items: [
                {
                    base_item_name: "Blue T-Shirt",
                    brand_name: "CoolBrand",
                    url: "/items/blue-tshirt",
                    thumbnail_url: "blue-tshirt.jpg",
                },
            ],
        },
        {
            name: "Work Closet",
            items: [
                {
                    base_item_name: "Grey Blazer",
                    brand_name: "SuitCo",
                    thumbnail_url: "grey-blazer.jpg",
                },
                {
                    base_item_name: "Black Pants",
                    brand_name: "PantsHouse",
                    thumbnail_url: "black-pants.jpg",
                },
            ],
        },
    ];

    const orders: Order[] = [
        {
            id: "FP-123102",
            date: "Mar 2, 2025",
            status: "Returned",
            total: "$59",
        },
        {
            id: "FP-122884",
            date: "Feb 18, 2025",
            status: "Completed",
            total: "$49",
        },
    ];

    return (
        <>
            <Hero />
            <div className="container">
                <div className="row g-4">
                    <div className="col-lg-4">
                        <ProfileCard user={user} />
                        <ClosetsCard closets={closets} />
                        <MembershipCard />
                    </div>

                    <div className="col-lg-8">
                        <CurrentShipmentCard />
                        <RecentOrdersCard orders={orders} />
                    </div>
                </div>
            </div>
        </>
    );
}

const domNode = document.getElementById("root")!;
const root = createRoot(domNode);
root.render(
    <React.StrictMode>
        <SWRConfig value={{ fetcher }}>
            <App />
        </SWRConfig>
    </React.StrictMode>,
);
