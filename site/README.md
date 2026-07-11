# `site/` — le site publié

**Ce dossier _est_ la racine du site en ligne.** GitHub Actions le téléverse tel quel vers
GitHub Pages : ce que vous voyez ici est exactement ce qui est servi.

**→ https://billfassinou.github.io/strongswan-manager/**

```
site/                        ← racine web
├── index.html               /                vitrine FR
├── en/index.html            /en/             vitrine EN
├── docs/                    /docs/           documentation FR (+ docs/en/ pour l'EN)
│   ├── index.html                            le visualiseur (fetch + rendu des .md)
│   └── *.md                                  les 20 pages, ×2 langues
├── assets/
│   ├── styles.css                            tokens de design (thème clair/sombre)
│   ├── logo.svg
│   └── og-image.png                          image de partage (1200×630)
├── robots.txt
├── sitemap.xml
└── .nojekyll                ← NE PAS SUPPRIMER
```

Statique intégral : pas de build, pas de bundler, **aucune requête externe** (ni police, ni CDN,
ni analytics). Le site doit rester affichable dans un réseau isolé.

## Deux pièges à connaître

1. **`.nojekyll` est indispensable.** Sans lui, GitHub ferait passer le dossier dans Jekyll, qui
   transformerait les fichiers `.md` en HTML — or le visualiseur de la documentation les
   `fetch()` en **Markdown brut**. La doc afficherait « Page introuvable ».
2. **Ouvrir `index.html` en `file://` casse la documentation.** Les navigateurs bloquent
   `fetch()` sur `file://`. Il faut passer par un serveur (voir ci-dessous).

## Prévisualiser en local

Servez **ce dossier** — il est la racine :

```bash
cd site && python3 -m http.server 8000
```

| Page | URL locale |
|---|---|
| Vitrine FR | http://localhost:8000/ |
| Vitrine EN | http://localhost:8000/en/ |
| Documentation FR | http://localhost:8000/docs/ |
| Documentation EN | http://localhost:8000/docs/en/ |

Les chemins locaux sont **identiques** à ceux de la production : ce qui marche ici marche en
ligne.

## Publication

Automatique : tout `push` sur `main` touchant `site/**` déclenche
[`.github/workflows/pages.yml`](../.github/workflows/pages.yml), qui téléverse ce dossier et
déploie. Aucune étape manuelle.

Prérequis, à faire **une seule fois** dans le dépôt GitHub :
*Settings → Pages → **Source = GitHub Actions***.

## Modifier

- **Contenu de la vitrine** : `index.html` et `en/index.html` sont indépendants — toute
  modification de l'un doit être **reportée dans l'autre**.
- **Contenu de la documentation** : éditez les `.md` dans `docs/` (et `docs/en/`). Pour ajouter
  une page, créez-la dans les deux langues puis déclarez-la dans le tableau `PARTS` de
  `docs/index.html` **et** de `docs/en/index.html`.
- **Design** : tout part des variables CSS en tête de `assets/styles.css`. Elles reprennent les
  tokens de l'interface réelle (`backend/web/src/styles.css`) — gardez les deux alignés.
- **Métadonnées SEO** : `canonical`, `hreflang` et Open Graph sont codés en dur dans les 4 shells
  HTML, et le sitemap liste 4 URLs. Si l'adresse du site change, mettez-les à jour ensemble
  (`grep -rl billfassinou.github.io site/`).
