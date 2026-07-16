# TP5 — Serveur MCP pour Mira

## Prérequis

- Go 1.26.2

## Installation

### 1. Lancer l'API Mira (depuis tp4/)

```bash
cd ../tp4
docker compose up -d # lance la bdd
go run ./cmd/api
```

L'API tourne sur `http://localhost:8080/api/v1`

### 2. Installer le serveur MCP

```bash
cd ../tp5
go mod tidy
go build ./cmd/mira-mcp
```

## Utilisation

### Claude Code

Copier `.mcp.json` dans votre configuration Claude Code (ou créer
`~/.claude/config.json` avec la section `mcpServers` correspondante) :

```bash
mkdir -p ~/.claude
cp .mcp.json ~/.claude/config.json
```

Redémarrer Claude Code. Le serveur `mira` apparaîtra alors dans la liste des outils
disponibles.

### Claude Desktop

1. Ouvrir `~/Library/Application\ Support/Claude/claude_desktop_config.json`
   (macOS) ou `%APPDATA%\Claude\claude_desktop_config.json` (Windows).

2. Ajouter :

```json
{
  "mcpServers": {
    "mira": {
      "command": "./mira-mcp",
      "env": {
        "MIRA_API_URL": "http://localhost:8080/api/v1"
      }
    }
  }
}
```

3. Redémarrer Claude Desktop.

## Outils disponibles

### `search_notes`

Recherche hybride (full-text + vecteur) dans les notes.

- `query` (string, requis) : requête de recherche
- `limit` (int, optionnel) : nombre max de résultats (défaut 10)

Exemple : "Retrouve ma note sur les channels Go"

### `get_note`

Récupère une note complète par ID.

- `id` (string, requis) : identifiant de la note

Retourne : titre, contenu, tags, résumé, score, statut d'enrichissement.

### `add_note`

Crée une nouvelle note.

- `title` (string, requis) : titre
- `content` (string, requis) : contenu
- `tags` ([]string, optionnel) : tags

Exemple : "Ajoute une note qui résume ce qu'on vient de faire"

### `list_recent_notes`

Liste les notes créées récemment.

- `limit` (int, optionnel) : nombre max de résultats (défaut 10)

## Architecture

```
tp5/
  cmd/mira-mcp/       binaire MCP
  internal/mira/      client HTTP + types
  .mcp.json           config pour Claude
  README.md
```

## Configuration

Variable d'environnement `MIRA_API_URL` (défaut `http://localhost:8080/api/v1`) :

```bash
MIRA_API_URL=http://localhost:8080/api/v1 go run ./cmd/mira-mcp
```

## Exemples de prompts

**Chercher une note :**
> "Retrouve ma note sur les channels Go et résume-la"

**Créer et chercher :**
> "Retrouve ma note sur les channels Go et ajoute une note résumant ce qu'on vient de faire"

**Enrichissement automatique :**
> "Crée une note sur les goroutines et quand tu vois que c'est enrichi, lis-la pour
> me montrer le résumé généré automatiquement"
