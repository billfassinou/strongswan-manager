import { createContext, useContext, useState, ReactNode, useCallback } from "react";

// --- Toasts ---
type Toast = { id: number; msg: string; kind: string };
const ToastCtx = createContext<(msg: string, kind?: string) => void>(() => {});

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const push = useCallback((msg: string, kind = "") => {
    const id = Date.now() + Math.random();
    setToasts((t) => [...t, { id, msg, kind }]);
    setTimeout(() => setToasts((t) => t.filter((x) => x.id !== id)), 2800);
  }, []);
  return (
    <ToastCtx.Provider value={push}>
      {children}
      <div id="toasts">
        {toasts.map((t) => (
          <div key={t.id} className={"toast " + t.kind}>
            {t.msg}
          </div>
        ))}
      </div>
    </ToastCtx.Provider>
  );
}
export const useToast = () => useContext(ToastCtx);

// --- Pill d'état ---
const LABELS: Record<string, string> = { u: "Actif", n: "Négociation", d: "Down", x: "Inconnu" };
export function statusClass(status: string): "u" | "n" | "d" | "x" {
  if (status === "up") return "u";
  if (status === "negotiating" || status === "installing") return "n";
  if (status === "down") return "d";
  return "x";
}
export function Pill({ status, text }: { status: string; text?: string }) {
  const c = statusClass(status);
  return (
    <span className={"pill " + c}>
      <span className={"dot " + c}></span> {text || LABELS[c]}
    </span>
  );
}

// --- Modale ---
export function Modal({
  title,
  onClose,
  children,
  footer,
}: {
  title: string;
  onClose: () => void;
  children: ReactNode;
  footer?: ReactNode;
}) {
  return (
    <div className="overlay" onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <div className="modal-head">
          <h3>{title}</h3>
          <button className="btn ghost x" onClick={onClose}>
            ✕
          </button>
        </div>
        <div className="modal-body">{children}</div>
        {footer && <div className="modal-foot">{footer}</div>}
      </div>
    </div>
  );
}

export function scoreColor(s: number): string {
  return s >= 85 ? "up" : s >= 65 ? "warn" : "down";
}
