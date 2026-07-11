-- Tunnel site-à-site géré des deux côtés : passerelle pair (optionnelle).
ALTER TABLE tunnels ADD COLUMN IF NOT EXISTS peer_gateway_id TEXT;
