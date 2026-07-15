# TP4 Mira

TP Go avec PostgreSQL, enrichissement automatique et recherche hybride.

## Démarrer PostgreSQL

```bash
docker compose up -d
```

## Lancer l'API

```bash
DATABASE_URL=postgres://mira:mira@localhost:5432/mira?sslmode=disable go run ./cmd/api
```

## Lancer la CLI

```bash
MIRA_API_URL=http://localhost:8080/api/v1 go run ./cmd/cli add "Titre" "Contenu" tag1,tag2
```

Les migrations se lancent au démarrage de l'API. La CLI passe par l'API HTTP.
