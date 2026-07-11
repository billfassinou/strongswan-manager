// Package domain contient les modèles métier, la validation et le scoring.
// Aucune dépendance à HTTP, à la base ou à VICI : c'est le cœur pur du produit.
package domain

import (
	"encoding/json"
	"time"
)

// Rôles RBAC (calqués sur la maquette front : admin/operator/auditor/viewer).
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleAuditor  = "auditor"
	RoleViewer   = "viewer"
)

// RoleCanWrite indique si un rôle a le droit d'effectuer des actions modifiantes.
func RoleCanWrite(role string) bool {
	return role == RoleAdmin || role == RoleOperator
}

// Types de topologie de tunnel autorisés (§14.2).
const (
	TypeSiteToSite  = "site-to-site"
	TypeRoadWarrior = "road-warrior"
	TypeHostToHost  = "host-to-host"
)

// Méthodes d'authentification (§16.2).
const (
	AuthPSK  = "psk"
	AuthCert = "cert"
	AuthEAP  = "eap"
)

// États d'une SA/tunnel.
const (
	StatusUp          = "up"
	StatusDown        = "down"
	StatusNegotiating = "negotiating"
	StatusUnknown     = "unknown"
	StatusInstalling  = "installing"
)

// User est un compte d'administration de la console.
type User struct {
	ID       string
	Identity string
	Role     string
	Enabled  bool
	PassHash string
}

// Gateway est une instance StrongSwan gérée.
type Gateway struct {
	ID        string
	Name      string
	Endpoint  string // "unix:/var/run/charon.vici" ou "tcp:host:port"
	Version   string
	Status    string
	CreatedAt time.Time
}

// Tunnel est la configuration d'une connexion IPsec (IKE_SA + CHILD_SA).
// Les champs suivent le dictionnaire de données §16.2 du cahier des charges.
type Tunnel struct {
	ID            string
	Name          string
	GatewayID     string
	PeerGatewayID *string // site-à-site géré des deux côtés : passerelle pair (optionnel)
	Type          string
	IKEVersion    int
	LocalAddr     string
	RemoteAddr    string
	LocalSubnets  []string
	RemoteSubnets []string
	AuthMethod    string
	SecretRef     *string
	CertRef       *string
	PeerCertRef   *string // certificat du pair (site-à-site par certificats)
	ProposalsIKE  []string
	ProposalsESP  []string
	PFS           bool
	Status        string
	SecurityScore int
	ConfigVersion int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ConfigVersion est un instantané horodaté d'une configuration de tunnel (GitOps interne).
type ConfigVersion struct {
	ID        string
	TunnelID  string
	N         int
	AuthorID  string
	Message   string
	Snapshot  []byte // JSON du Tunnel au moment de l'application
	CreatedAt time.Time
}

// AuditEntry est une entrée immuable du journal d'audit (append-only, chaînée).
type AuditEntry struct {
	ID            string
	ActorID       string
	Action        string
	Target        string
	Timestamp     time.Time
	IntegrityHash string
	PrevHash      string
}

// Types de secrets pris en charge (EF-05).
const (
	SecretPSK   = "psk"
	SecretEAP   = "eap"
	SecretXAuth = "xauth"
)

// Secret est un secret chiffré au repos (la valeur en clair ne quitte jamais le coffre).
type Secret struct {
	ID       string
	Name     string
	Type     string
	UsedBy   string
	EncValue []byte // valeur chiffrée (AES-GCM) ; jamais exposée par l'API
}

// CA est l'autorité de certification interne (clé privée chiffrée au repos).
type CA struct {
	ID        string
	Name      string
	CertPEM   []byte
	KeyEnc    []byte
	CRLNumber int64
	CRLPEM    []byte
}

// RevokedCert identifie un certificat révoqué pour la génération de CRL.
type RevokedCert struct {
	Serial    string
	RevokedAt time.Time
}

// Types et états de certificats (§16.2).
const (
	CertKindServer = "server"
	CertKindClient = "client"
	CertValid      = "valid"
	CertRevoked    = "revoked"
	CertExpired    = "expired"
)

// Certificate est un certificat X.509 émis par la CA interne.
type Certificate struct {
	ID        string
	Name      string
	CN        string
	Kind      string
	Serial    string
	Status    string
	NotBefore time.Time
	NotAfter  time.Time
	CertPEM   []byte
	KeyEnc    []byte // clé privée chiffrée ; jamais exposée par l'API
}

// ConfigItem est une entité de configuration générique (pool, RADIUS, politique,
// autorité, utilisateur VPN, règle d'alerte, paramètres du démon…) : ses champs sont
// portés en JSON, ce qui mutualise le stockage et le CRUD entre modules.
type ConfigItem struct {
	ID   string
	Kind string
	Name string
	Data json.RawMessage
}

// SAState est l'état temps réel d'une SA remonté par VICI (list-sas).
type SAState struct {
	Name    string
	Status  string // up / negotiating / down
	BytesIn uint64
	BytesOut uint64
}
