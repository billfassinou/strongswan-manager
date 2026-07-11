package httpapi

import (
	"context"
	"encoding/json"

	"strongswan-manager/internal/domain"
)

// Interfaces minimales dont dépendent les handlers. Les repositories concrets du
// paquet store les satisfont automatiquement ; les tests fournissent des fakes en
// mémoire, ce qui rend la couche HTTP testable sans PostgreSQL.

type usersStore interface {
	GetByIdentity(ctx context.Context, identity string) (*domain.User, error)
}

type gatewaysStore interface {
	Get(ctx context.Context, id string) (*domain.Gateway, error)
	List(ctx context.Context) ([]domain.Gateway, error)
}

type tunnelsStore interface {
	List(ctx context.Context) ([]domain.Tunnel, error)
	Get(ctx context.Context, id string) (*domain.Tunnel, error)
	Create(ctx context.Context, t *domain.Tunnel) error
	Update(ctx context.Context, t *domain.Tunnel) error
	Delete(ctx context.Context, id string) error
}

type versionsStore interface {
	Create(ctx context.Context, v *domain.ConfigVersion) error
	ListByTunnel(ctx context.Context, tunnelID string) ([]domain.ConfigVersion, error)
	Get(ctx context.Context, tunnelID string, n int) (*domain.ConfigVersion, error)
}

type auditStore interface {
	Append(ctx context.Context, actorID, action, target string) error
	List(ctx context.Context, limit int) ([]domain.AuditEntry, error)
}

type secretsStore interface {
	Create(ctx context.Context, s *domain.Secret) error
	GetByName(ctx context.Context, name string) (*domain.Secret, error)
	List(ctx context.Context) ([]domain.Secret, error)
	Delete(ctx context.Context, id string) error
}

// cipher chiffre/déchiffre les valeurs de secrets au repos.
type cipher interface {
	Encrypt(plain []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

type caStore interface {
	Get(ctx context.Context) (*domain.CA, error)
	UpdateCRL(ctx context.Context, id string, number int64, crlPEM []byte) error
}

type certsStore interface {
	Create(ctx context.Context, c *domain.Certificate) error
	GetByName(ctx context.Context, name string) (*domain.Certificate, error)
	List(ctx context.Context) ([]domain.Certificate, error)
	Revoke(ctx context.Context, id string) error
	ListRevoked(ctx context.Context) ([]domain.RevokedCert, error)
}

type configStore interface {
	Create(ctx context.Context, c *domain.ConfigItem) error
	List(ctx context.Context, kind string) ([]domain.ConfigItem, error)
	Get(ctx context.Context, id string) (*domain.ConfigItem, error)
	Update(ctx context.Context, id, name string, data json.RawMessage) error
	Delete(ctx context.Context, id string) error
}
