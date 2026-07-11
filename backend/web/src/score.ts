// Port TypeScript du score de sécurité (aligné sur le backend domain.ScoreTunnel
// et la maquette app.html), pour un retour immédiat dans l'éditeur.

export const ALGOS = [
  "aes256gcm16",
  "aes128gcm16",
  "aes256",
  "3des",
  "sha384",
  "sha256",
  "md5",
  "ecp384",
  "ecp256",
  "modp2048",
  "modp1024",
  "mlkem768",
  "mlkem512",
];
export const WEAK = ["3des", "des", "md5", "modp1024", "modp768"];

export interface ScoreResult {
  score: number;
  notes: string[];
}

export function scoreTunnel(t: {
  ike_version: number;
  pfs: boolean;
  proposals_ike: string[];
  proposals_esp: string[];
}): ScoreResult {
  let s = 100;
  const notes: string[] = [];
  const j = [...(t.proposals_ike || []), ...(t.proposals_esp || [])].join(" ").toLowerCase();

  if (t.ike_version === 1) {
    s -= 42;
    notes.push("IKEv1 obsolète — migrer vers IKEv2");
  }
  if (/3des/.test(j) || /(^|[-_\s])des([-_\s]|$)/.test(j)) {
    s -= 28;
    notes.push("Chiffrement 3DES/DES faible");
  }
  if (/md5/.test(j)) {
    s -= 18;
    notes.push("Empreinte MD5 faible");
  }
  if (/modp1024|modp768/.test(j)) {
    s -= 16;
    notes.push("Groupe Diffie-Hellman faible (modp1024)");
  }
  if (!t.pfs) {
    s -= 10;
    notes.push("Perfect Forward Secrecy désactivé");
  }
  if (!/mlkem/.test(j)) {
    s -= 6;
    notes.push("Pas de préparation post-quantique (ML-KEM)");
  }
  return { score: Math.max(5, Math.min(100, s)), notes };
}

export function scoreColor(s: number): string {
  return s >= 85 ? "up" : s >= 65 ? "warn" : "down";
}
