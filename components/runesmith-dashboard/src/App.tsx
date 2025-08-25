import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Github, Linkedin, Flame, Snowflake, Sparkles, ExternalLink, Info, X, Moon, Sun } from "lucide-react";

// ---------- Types from backend ----------
export type ArtifactStatus =
    | "Scheduled"
    | "Requeued"
    | "Failed"
    | "Enchanting"
    | "Completed"
    | string;

export interface Artifact {
    ID: number;
    ItemID: number;
    ItemName: string;
    TaskID: string;
    CreatedAt: string; // RFC3339 from Go
    UpdatedAt: string;
    Status: ArtifactStatus;
}

export interface NodeStatus {
    Name: string;
    Available: number;
    Allocated: number;
    Healthy: boolean;
    RunningJobs: number;
}

// items schema may vary; keep it flexible
export interface ItemRequirements { Fire?: number; Frost?: number; Arcane?: number }
export interface Item { ID: number; Name: string; Tier?: string; Requirements?: ItemRequirements; Priority?: number }

// ---------- Config ----------
const API_BASE = ""; // keep relative; dev proxy should map /api to backend
const API = `${API_BASE}/api/v1`;

// ---------- Utilities ----------
function cx(...a: (string | false | null | undefined)[]) { return a.filter(Boolean).join(" "); }

type Energy = "fire" | "frost" | "arcane";
const ENERGY_HEX: Record<Energy, string> = {
    fire: "#AA4203",
    frost: "#007FFF",
    arcane: "#bc04bf",
};

const energyFromNodeName = (name: string): Energy => {
    const n = (name || "").toLowerCase();
    if (/(\(fire\)|fire)/.test(n)) return "fire";
    if (/(\(frost\)|frost)/.test(n)) return "frost";
    if (/(\(arcane\)|arcane)/.test(n)) return "arcane";
    return "undefined" as "fire" | "frost" | "arcane";
};

const energyIcon = (e: Energy, cls = "") => {
    switch (e) {
        case "fire": return <Flame className={cls} />;
        case "frost": return <Snowflake className={cls} />;
        case "arcane": return <Sparkles className={cls} />;
        default: return <Info className={cls} />;
    }
};

// Normalize /items payloads from different shapes
function normalizeItemsPayload(j: any): Item[] {
    const arr: any[] = Array.isArray(j) ? j : j.items || j.artifacts || [];
    return (arr || []).map((x: any) => {
        const id = x.ID ?? x.Id ?? x.id;
        const name = x.Name ?? x.name;
        const tier = x.Tier ?? x.tier;
        const req = x.Requirements ?? x.requirements ?? {};
        const fire = req.Fire ?? req.fire ?? 0;
        const frost = req.Frost ?? req.frost ?? 0;
        const arcane = req.Arcane ?? req.arcane ?? 0;
        const priority = x.Priority ?? x.priority;
        const it: Item = {
            ID: Number(id),
            Name: String(name),
            Tier: tier,
            Requirements: { Fire: Number(fire), Frost: Number(frost), Arcane: Number(arcane) },
            Priority: priority,
        };
        return it;
    });
}

const rarityClass = (id?: number) => {
    if (!id) return "";
    if (id >= 31) return "text-orange-600 dark:text-orange-600 font-medium"; // Legendary
    if (id >= 21) return "text-blue-600 dark:text-blue-400 font-medium";     // Epic
    if (id >= 11) return "text-green-600 dark:text-green-500 font-medium";    // Rare
    return "";                                                                // Common
};


// ---------- Toasts ----------
// Theme (dark mode) helper
const useDarkMode = () => {
    const [dark, setDark] = useState<boolean>(() => {
        try {
            const fromStorage = localStorage.getItem("theme");
            if (fromStorage === "dark") return true;
            if (fromStorage === "light") return false;
        } catch {}
        return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
    });
    useEffect(() => {
        const root = document.documentElement;
        if (dark) { root.classList.add("dark"); localStorage.setItem("theme", "dark"); }
        else { root.classList.remove("dark"); localStorage.setItem("theme", "light"); }
    }, [dark]);
    return { dark, toggle: () => setDark(d => !d) };
};

// ---------- Toasts ----------
interface Toast { id: number; msg: string }
const useToasts = () => {
    const [toasts, setToasts] = useState<Toast[]>([]);
    const seq = useRef(0);
    const push = useCallback((msg: string) => {
        const id = ++seq.current; setToasts(t => [{ id, msg }, ...t]);
        setTimeout(() => setToasts(t => t.filter(x => x.id !== id)), 2000);
    }, []);
    return { toasts, push };
};

const ToastStack: React.FC<{ toasts: Toast[] }> = ({ toasts }) => (
    <div className="fixed right-3 top-3 z-50 flex flex-col gap-2">
        {toasts.map((t) => {
            const isError = /failed|error|rate\s*limited|http\s*(4\d{2}|5\d{2})/i.test(t.msg);
            return (
                <div
                    key={t.id}
                    className={cx(
                        "rounded-2xl text-white shadow-lg px-4 py-2 text-sm backdrop-blur",
                        isError ? "bg-red-600 dark:bg-red-500" : "bg-black/80"
                    )}
                >
                    {t.msg}
                </div>
            );
        })}
    </div>
);

// ---------- Modals ----------
const Modal: React.FC<{ open: boolean; onClose: () => void; title?: string; children: React.ReactNode }>
    = ({ open, onClose, title, children }) => {
    if (!open) return null;
    return (
        <div className="fixed inset-0 z-40 flex items-center justify-center">
            <div className="absolute inset-0 bg-black/40" onClick={onClose} />
            <div className="relative z-50 w-[min(92vw,900px)] max-h-[90vh] overflow-auto rounded-2xl bg-white dark:bg-slate-900 p-6 shadow-2xl">
                <div className="mb-3 flex items-center justify-between">
                    <h3 className="text-lg font-semibold">{title}</h3>
                    <button onClick={onClose} className="rounded-full p-1 hover:bg-black/5 dark:hover:bg-white/10"><X size={18} /></button>
                </div>
                {children}
            </div>
        </div>
    );
};

// ---------- Header & Footer ----------
const Header: React.FC<{ onOpenAbout: () => void; onOpenItems: () => void; onToggleTheme: () => void; dark: boolean }>
    = ({ onOpenAbout, onToggleTheme, dark }) => (
    <header className="sticky top-0 z-30 bg-white/80 dark:bg-slate-900/80 backdrop-blur border-b border-slate-200 dark:border-slate-700">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
            <div className="flex items-center gap-3">
                <a href="/" className="flex items-center gap-2">
                    <img src="/runesmith.png" alt="Runesmith" className="h-8 w-8 rounded-xl border dark:border-slate-700 object-cover" />
                    <span className="text-sm font-semibold tracking-tight text-slate-800 dark:text-slate-200">Runesmith</span>
                </a>
                <nav className="hidden md:flex items-center gap-6 text-sm text-slate-700 dark:text-slate-300">
                    <a
                        className="inline-flex items-center gap-1 rounded-full px-3 py-1.5
                       bg-gradient-to-r from-rose-500 to-orange-500 text-white
                       shadow-sm ring-1 ring-inset ring-white/10 transition
                       hover:shadow-md focus-visible:outline-none
                       focus-visible:ring-2 focus-visible:ring-rose-400/70"
                        href="https://chat.skypiea-ai.xyz"
                        target="_blank"
                        rel="noreferrer"
                        aria-label="Visit Skypiea (opens in a new tab)">Skypiea <ExternalLink className="-mt-0.5" size={14} />
                    </a>
                    <button className="hover:text-black dark:hover:text-white" onClick={onOpenAbout}>About</button>
                </nav>
            </div>
            <div className="flex items-center gap-1 sm:gap-2 text-slate-700 dark:text-slate-300">
                <button onClick={onToggleTheme} className="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/10" aria-label="Toggle dark mode">
                    {dark ? <Sun size={18} /> : <Moon size={18} />}
                </button>
                <a className="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/10" href="https://github.com/fukaraca/runesmith" target="_blank" rel="noreferrer" aria-label="GitHub"><Github size={18} /></a>
                <a className="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/10" href="https://www.linkedin.com/in/fukaraca" target="_blank" rel="noreferrer" aria-label="LinkedIn"><Linkedin size={18} /></a>
            </div>
        </div>
    </header>
);

const Footer: React.FC = () => (
    <footer className="mt-12 border-t border-slate-200 dark:border-slate-700">
        <div className="mx-auto max-w-6xl px-4 py-6 text-xs text-slate-500 dark:text-slate-400">
            Runesmith demo — Kubernetes CRDs & Operator, Kueue priority scheduling, custom Device Plugin with virtual resources(manawell.io/fire · manawell.io/frost · manawell.io/arcane), Go backend & React dashboard, microservices, deployed on AWS EC2 (kubeadm).
        </div>
    </footer>
);

// ---------- Node Card ----------
const NodeCard: React.FC<{ s: NodeStatus }>= ({ s }) => {
    const energy = energyFromNodeName(s.Name);
    const color = ENERGY_HEX[energy];
    const total = s.Allocated + s.Available;
    const pct = total > 0 ? Math.round((s.Allocated / total) * 100) : 0;
    const barStyle: React.CSSProperties = { width: `${pct}%`, backgroundColor: color };

    return (
        <div className="rounded-2xl border border-slate-200 dark:border-slate-700 p-4 shadow-sm bg-white dark:bg-slate-900">
            <div className="mb-2 flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <div style={{ color }}>{energyIcon(energy, "h-5 w-5")}</div>
                    <div className="font-medium">{s.Name}</div>
                </div>
                <div className={cx("rounded-full px-2 py-0.5 text-xs", s.Healthy ? "bg-emerald-50 text-emerald-700" : "bg-rose-50 text-rose-700")}>{s.Healthy ? "Healthy" : "Unhealthy"}</div>
            </div>
            <div className="mt-2">
                <div className="mb-1 flex items-center justify-between text-xs text-slate-600 dark:text-slate-400">
                    <span>allocated: <span style={{ color }} className="font-semibold">{s.Allocated}</span></span>
                    <span>available: <span style={{ color }} className="font-semibold">{s.Available}</span></span>
                </div>
                <div className="h-3 w-full overflow-hidden rounded-full bg-slate-100 dark:bg-slate-800"><div className="h-full transition-all" style={barStyle} /></div>
            </div>
            <div className="mt-2 text-xs text-slate-600 dark:text-slate-400">Status: {s.RunningJobs} job(s) running.</div>
        </div>
    );
};

// ---------- Lists ----------
const ArtifactsTable: React.FC<{ title: string; artifacts: Artifact[] }> = ({ title, artifacts }) => (
    <div className="rounded-2xl border border-slate-200 dark:border-slate-700 p-4 shadow-sm bg-white dark:bg-slate-900">
        <div className="mb-3 text-sm font-semibold">{title}</div>
        <div className="overflow-x-auto">
            <table className="min-w-full text-sm">
                <thead className="text-left text-slate-500 dark:text-slate-400">
                <tr>
                    <th className="px-2 py-1">ID</th>
                    <th className="px-2 py-1">ItemID</th>
                    <th className="px-2 py-1 w-[22rem] md:w-[30rem]">ItemName</th>
                    <th className="px-2 py-1">Status</th>
                    <th className="px-2 py-1">Created</th>
                </tr>
                </thead>
                <tbody>
                {artifacts.map((a) => (
                    <tr key={a.ID} className="border-t border-slate-100 dark:border-slate-800">
                        <td className="px-2 py-1 font-mono">{a.ID}</td>
                        <td className="px-2 py-1">{a.ItemID}</td>
                        <td className="px-2 py-1 w-[22rem] md:w-[30rem]">
                <span
                    className={`block truncate ${rarityClass(a.ItemID ?? 0)}`}
                    title={a.ItemName ?? ""}
                >
                  {a.ItemName}
                </span>
                        </td>
                        <td className="px-2 py-1">{a.Status}</td>
                        <td className="px-2 py-1 text-slate-500 dark:text-slate-400">
                            {new Date(a.CreatedAt).toLocaleString()}
                        </td>
                    </tr>
                ))}
                {artifacts.length === 0 && (
                    <tr>
                        <td colSpan={5} className="px-2 py-3 text-center text-slate-400 dark:text-slate-500">
                            Nothing here yet.
                        </td>
                    </tr>
                )}
                </tbody>
            </table>
        </div>
    </div>
);

// ---------- Main App ----------
const App: React.FC = () => {
    const { dark, toggle } = useDarkMode();
    const { toasts, push } = useToasts();

    const [aboutOpen, setAboutOpen] = useState(false);
    const [itemsOpen, setItemsOpen] = useState(false);

    const [nodeStatuses, setNodeStatuses] = useState<NodeStatus[]>([]);
    const [pending, setPending] = useState<Artifact[]>([]);
    const [completed, setCompleted] = useState<Artifact[]>([]);
    const [items, setItems] = useState<Item[]>([]);

    const hasRunning = useMemo(() => nodeStatuses.some(n => n.RunningJobs > 0), [nodeStatuses]);
    const hasPending = pending.length > 0;

    const fetchStatus = useCallback(async () => {
        try {
            const r = await fetch(`${API}/status`);
            const j = await r.json();
            const s: NodeStatus[] = Array.isArray(j) ? j : j.status;
            setNodeStatuses(Array.isArray(s) ? s : []);
        } catch { push("Failed to fetch status"); }
    }, [push]);

    const fetchArtifacts = useCallback(async () => {
        try {
            const [p, c] = await Promise.all([
                fetch(`${API}/artifacts?completed=false`).then(r => r.json()),
                fetch(`${API}/artifacts?completed=true`).then(r => r.json()),
            ]);
            const norm = (x: any): Artifact[] => (Array.isArray(x) ? x : x.artifacts) || [];
            const sortByCreated = (arr: Artifact[]) => [...arr].sort((a, b) => new Date(b.CreatedAt).getTime() - new Date(a.CreatedAt).getTime());
            const sortByUpdated = (arr: Artifact[]) => [...arr].sort((a, b) => new Date(b.UpdatedAt).getTime() - new Date(a.UpdatedAt).getTime());
            setPending(sortByCreated(norm(p)));
            setCompleted(sortByUpdated(norm(c)));
        } catch { push("Failed to fetch artifacts"); }
    }, [push]);

    const fetchItems = useCallback(async () => {
        try {
            const r = await fetch(`${API}/items`);
            const j = await r.json();
            const list: Item[] = normalizeItemsPayload(j);
            setItems(list);
        } catch { push("Failed to fetch items"); }
    }, [push]);

    const forge = useCallback(async () => {
        try {
            const r = await fetch(`${API}/forge`, { method: "POST" });
            if (r.status === 429) {
                const ra = r.headers.get("Retry-After");
                push(`Rate limited (HTTP 429).${ra ? ` Retry after: ${ra}.` : ""}`);
                return;
            }
            if (!r.ok) {
                push(`Forge failed (HTTP ${r.status}).`);
                return;
            }

            const j = await r.json();
            const name = j?.job_name || "unknown";
            push(`Job ${name} created`);
            await Promise.all([fetchArtifacts(), fetchStatus()]);
        } catch {
            push("Forge request failed");
        }
    }, [fetchArtifacts, fetchStatus, push]);

    useEffect(() => { fetchStatus(); fetchArtifacts(); }, [fetchStatus, fetchArtifacts]);

    useEffect(() => {
        if (!(hasRunning || hasPending)) return; // poll only if there is activity
        const id = setInterval(() => { fetchStatus(); fetchArtifacts(); }, 1000);
        return () => clearInterval(id);
    }, [hasRunning, hasPending, fetchStatus, fetchArtifacts]);

    useEffect(() => { if (itemsOpen) fetchItems(); }, [itemsOpen, fetchItems]);

    return (
        <div className="min-h-screen bg-slate-50 text-slate-800 dark:bg-slate-950 dark:text-slate-200 flex flex-col">
            <Header onOpenAbout={() => setAboutOpen(true)} onOpenItems={() => setItemsOpen(true)} onToggleTheme={toggle} dark={dark} />

            <main className="mx-auto max-w-6xl px-4 flex-1">
                <div className="flex flex-wrap items-center gap-3 py-4">
                    <button onClick={forge} className="rounded-2xl px-4 py-2 text-sm font-semibold shadow-sm transition text-white bg-orange-900 hover:bg-orange-700">Forge</button>
                    <button onClick={() => setItemsOpen(true)} className="text-sm underline underline-offset-4">List Of Possible Artifacts</button>
                </div>

                <section className="mb-8">
                    <div className="mb-3 text-sm font-semibold">Foundry(~nodes)</div>
                    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
                        {nodeStatuses.map(ns => <NodeCard key={ns.Name} s={ns} />)}
                        {nodeStatuses.length === 0 && (
                            <div className="rounded-2xl border border-dashed border-slate-300 p-6 text-center text-slate-400 dark:text-slate-500">No node data yet.</div>
                        )}
                    </div>
                </section>

                <section className="mb-6">
                    <ArtifactsTable title="Artifacts in Production" artifacts={pending} />
                </section>
                <section className="mb-6">
                    <ArtifactsTable title="Completed Orders" artifacts={completed} />
                </section>
            </main>

            <Footer />

            <Modal open={aboutOpen} onClose={() => setAboutOpen(false)} title="About Runesmith">
                <div className="prose prose-slate dark:prose-invert max-w-none">
                    <p>
                        Runesmith is a Kubernetes-native demo that “forges” magical artifacts using a virtual resource
                        called <em>Mana</em>. It’s built to showcase CRDs &amp; Operators, a custom Device Plugin,
                        and priority scheduling with Kueue, no expensive real hardware required.
                    </p>

                    <h4>What’s under the hood</h4>
                    <ul>
                        <li>
                            <strong>Custom Device Plugin</strong> exposes three resources:
                            <code className="mx-1">manawell.io/fire</code>,
                            <code className="mx-1">manawell.io/frost</code>,
                            <code className="mx-1">manawell.io/arcane</code>.
                            Each essence consumes 1 Mana on its matching node while running. Source of truth for resources.
                        </li>
                        <li>
                            <strong>CRD + Operator</strong> (<code>Enchantment</code>) defines an artifact order and manages
                            the lifecycle of the Jobs it requires.
                        </li>
                        <li>
                            <strong>Kueue</strong> enforces priorities and can queue or preempt workloads when Mana is scarce.
                        </li>
                        <li>
                            <strong>Backend &amp; UI</strong> (Go + React) submit orders, watch status, and visualize Mana and progress.
                        </li>
                    </ul>

                    <h4>Example: end-to-end forging</h4>
                    <pre className="whitespace-pre-wrap"><code>{`id: 38
name: "God-slayer Elemental Blade"
tier: "Legendary"
requirements:
  fire: 6
  frost: 5
  arcane: 4
priority: 2`}</code></pre>

                    <ol className="list-decimal pl-5">
                        <li>The UI requests a random item from the catalog.</li>
                        <li>The backend creates an <code>Enchantment</code> (CRD) representing that order.</li>
                        <li>The Operator notices the new Enchantment and spawns the required Jobs:
                            <b> 6x</b> represented as time consuming job on <code className="mx-1">Fire</code> tainted node.</li>
                        <li>Kueue schedules those Jobs based on available Mana; higher tiers (Legendary) carry higher priority and can preempt queued lower-tier work.</li>
                        <li>Each node’s Device Plugin advertises its Mana inventory to kubelet(eg Nvidia GPU nodes vs AMD GPU nodes); Jobs consume the matching resource while running.</li>
                        <li>As Jobs finish, Mana is released; Enchantment(CR) status updated; the backend tracks completion and the UI updates live.</li>
                        <li>When all required essences are done, the Enchantment is marked <code>Completed</code>.</li>
                    </ol>

                    <h4>How to read the UI</h4>
                    <ul>
                        <li><strong>Forge</strong>: submit a new Enchantment (random item).</li>
                        <li><strong>List of possible items</strong>: view the catalog (requirements per energy).</li>
                        <li><strong>Artifacts</strong>: see live orders and statuses
                            (<code>Scheduled</code>, <code>Enchanting</code>, <code>Requeued</code>, <code>Failed</code>, <code>Completed</code>).</li>
                        <li><strong>Nodes</strong>: real-time availability and allocation per node and energy type.</li>
                    </ul>

                    <p className="text-xs">
                        Purposefully a demo: it doesn’t solve a business problem; it shows how to build Kubernetes-native systems with
                        CRDs/Operators, a Device Plugin, and Kueue while keeping everything observable and reproducible.
                    </p>
                </div>
            </Modal>

            <Modal open={itemsOpen} onClose={() => setItemsOpen(false)} title="List of Possible Items">
                <div className="overflow-x-auto">
                    <table className="min-w-full text-sm">
                        <thead className="text-left text-slate-500 dark:text-slate-400">
                        <tr>
                            <th className="px-2 py-1">ID</th>
                            <th className="px-2 py-1">Name</th>
                            <th className="px-2 py-1">Tier</th>
                            <th className="px-2 py-1">Req (Fire / Frost / Arcane)</th>
                        </tr>
                        </thead>
                        <tbody>
                        {items.map((it) => (
                            <tr key={it.ID} className="border-t border-slate-100 dark:border-slate-800">
                                <td className="px-2 py-1">{it.ID}</td>
                                <td className={`px-2 py-1`}>
                                    <span className={`block truncate ${rarityClass(it.ID)}`} title={it.Name ?? ""}>
                                        {it.Name}
                                    </span>
                                </td>
                                <td className="px-2 py-1">{it.Tier || "-"}</td>
                                <td className="px-2 py-1">
                                    <span style={{ color: ENERGY_HEX.fire }} className="font-semibold">{it.Requirements?.Fire ?? 0}</span>
                                    {" / "}
                                    <span style={{ color: ENERGY_HEX.frost }} className="font-semibold">{it.Requirements?.Frost ?? 0}</span>
                                    {" / "}
                                    <span style={{ color: ENERGY_HEX.arcane }} className="font-semibold">{it.Requirements?.Arcane ?? 0}</span>
                                </td>
                            </tr>
                        ))}
                        {items.length === 0 && (
                            <tr><td colSpan={4} className="px-2 py-3 text-center text-slate-400 dark:text-slate-500">No items to show.</td></tr>
                        )}
                        </tbody>
                    </table>
                </div>
            </Modal>

            <ToastStack toasts={toasts} />
        </div>
    );
};

export default App;
