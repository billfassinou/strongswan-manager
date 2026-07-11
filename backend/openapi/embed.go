// Package openapi embarque la spécification OpenAPI et la page de documentation.
package openapi

import _ "embed"

// Spec contient le contrat OpenAPI (source de vérité du §10), servi sur /api/v1/openapi.yaml.
//
//go:embed openapi.yaml
var Spec []byte

// DocsHTML est une page Swagger UI minimale (chargée depuis un CDN en développement ;
// à vendorer pour un déploiement air-gapped).
const DocsHTML = `<!doctype html>
<html lang="fr">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>StrongSwan Manager — API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({ url: '/api/v1/openapi.yaml', dom_id: '#swagger-ui' });
  </script>
</body>
</html>`
