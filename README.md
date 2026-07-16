[![CI](https://github.com/Zastial/ESGI-Go-exercices/actions/workflows/ci.yml/badge.svg)](https://github.com/Zastial/ESGI-Go-exercices/actions/workflows/ci.yml)

# Projet Mira - Parcours Go ESGI

Auteur : Alexandre CAROL


### Fonctionnalités

**CLI mira locale** — `tp4/cmd/cli/`
- Commandes dispo : `mira add`, `mira list`, `mira search`
- Pour tester :
  ```bash
  cd tp4

  docker compose up -d # lancer la bdd

  MIRA_API_URL=http://localhost:8080/api/v1 go run ./cmd/cli add "Note test" "Contenu" tag1,tag2
  MIRA_API_URL=http://localhost:8080/api/v1 go run ./cmd/cli list
  MIRA_API_URL=http://localhost:8080/api/v1 go run ./cmd/cli search "tag1"
  ```

**API mira v1** — `tp4/cmd/api/`
- Pour tester :
  ```bash
  # Créer une note (doit retourner 201)
  curl -X POST http://localhost:8080/api/v1/notes -H "Content-Type: application/json" -d '{"title":"Test","content":"Contenu","tags":["go"]}'

  # Lister les notes
  curl http://localhost:8080/api/v1/notes

  # Récupérer une note par ID
  curl http://localhost:8080/api/v1/notes/ID

  # Modifier une note
  curl -X PATCH http://localhost:8080/api/v1/notes/ID -H "Content-Type: application/json" -d '{"title":"Nouveau titre"}'

  # Supprimer une note
  curl -X DELETE http://localhost:8080/api/v1/notes/ID
  
  # Rechercher
  curl "http://localhost:8080/api/v1/search?q=golang"
  ```

**Middleware request ID et validation**
- Pour tester :
  ```bash
  # Vérifier la structure de réponse
  curl -s http://localhost:8080/api/v1/notes | jq '.'

  # Envoyer un payload invalide (doit retourner 400)
  curl -X POST http://localhost:8080/api/v1/notes -H "Content-Type: application/json" -d '{"invalid":"payload"}'
  ```

**PostgreSQL + Migrations**

- Test :
  ```bash
  docker exec -it tp4-db-1 psql -U mira -d mira -c "\dt"

  docker exec -it tp4-db-1 psql -U mira -d mira -c "SELECT id, title, enrichment_status FROM notes LIMIT 5;"
  ```

**Enrichissement automatique**

- Pour tester :
  ```bash
  # Créer une note et observer le changement de statut
  curl -X POST http://localhost:8080/api/v1/notes -H "Content-Type: application/json" -d '{"title":"Test","content":"Contenu"}'

  # Attendre 1-2 sec
  curl -s http://localhost:8080/api/v1/notes | jq '.data[].enrichment_status'
  ```

**Recherche hybride (full-text + vectorielle)**

- Tester :
  ```bash
  # Ajouter des notes avec du contenu
  curl -X POST http://localhost:8080/api/v1/notes -H "Content-Type: application/json" -d '{"title":"Goroutines","content":"Les goroutines sont les threads légers du Go"}'

  # Rechercher
  curl "http://localhost:8080/api/v1/search?q=goroutines"
  ```

**Serveur MCP** — `tp5/cmd/mira-mcp/`

Pour tester : Se référer au README du tp5

## Structure du projet

```
.
├── README.md
├── go-warmup/
├── tp1-mira/
├── tp2/
├── tp3/
├── tp4/
│   ├── cmd/
│   │   ├── cli/
│   │   └── api/
│   ├── internal/
│   │   ├── core/
│   │   ├── config/
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   ├── response/
│   │   │   └── router.go
│   │   ├── service/
│   │   ├── store/
│   │   └── migrate
│   ├── migrations/
│   ├── docker-compose.yml
│   ├── go.mod / go.sum
│   └── README.md
│
└── tp5/
    ├── cmd/mira-mcp/           # MCP server
    │   ├── main.go
    │   └── handlers.go          # Tools : search, get, add, list
    ├── internal/mira/
    │   ├── client.go
    │   └── note.go
    ├── .mcp.json               # Config Claude
    ├── go.mod / go.sum
    ├── mira-mcp                # binaire compilée
    └── README.md
```

## Guide de démarrage

### Prérequis globaux

- **Go** 1.26.2+
- **Docker**

Tester tout le projet (API + CLI + MCP) :
```bash
# Terminal 1 : API
cd tp4
docker compose up -d
DATABASE_URL=postgres://mira:mira@localhost:5432/mira?sslmode=disable go run ./cmd/api

# Terminal 2 : CLI
cd tp4
MIRA_API_URL=http://localhost:8080/api/v1 go run ./cmd/cli add "Titre" "Contenu" tag1,tag2

# Terminal 3 : MCP Server
cd tp5
go build ./cmd/mira-mcp
# Puis configurer dans Claude Code / Claude Desktop
```

## Tests

### Tests unitaires & intégration (À faire)
```bash
cd tp4
go test ./...
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Claude Code / Claude Desktop                           │
│  └─ Tools MCP (search, get, add, list)                 │
└─────────────────────┬───────────────────────────────────┘
                      │ stdio
                      ▼
┌─────────────────────────────────────────────────────────┐
│  TP5 : mira-mcp (Serveur MCP)                          │
│  └─ Appelle l'API HTTP                                 │
└─────────────────────┬───────────────────────────────────┘
                      │ HTTP Client
                      ▼
┌─────────────────────────────────────────────────────────┐
│  TP4 : API REST Mira v1                                │
│  ├─ POST /api/v1/notes (201)                           │
│  ├─ GET /api/v1/notes (200)                            │
│  ├─ GET /api/v1/notes/{id} (200/404)                   │
│  ├─ PATCH /api/v1/notes/{id} (200)                     │
│  ├─ DELETE /api/v1/notes/{id} (204)                    │
│  ├─ GET /api/v1/search?q=... (200)                     │
│  └─ Middleware : Request ID, Logging                   │
└─────┬──────────────────────────────────┬─────────────────┘
      │ Queue & Workers                  │ HTTP Client
      ▼                                   ▼
┌──────────────────────┐  ┌──────────────────────────────┐
│ Enrichissement Pool  │  │  CLI : mira add/list/search  │
│ (3 workers, 4s)     │  │  └─ Appelle l'API HTTP      │
└──────────┬───────────┘  └──────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────┐
│  PostgreSQL 15 + pgvector                              │
│  ├─ Table notes (id, title, content, tags, summary)    │
│  ├─ Index GIN pour full-text                           │
│  ├─ Vector embeddings pour recherche vectorielle       │
│  ├─ enrichment_status (pending/done/failed)            │
│  └─ created_at, updated_at                             │
└─────────────────────────────────────────────────────────┘
```
