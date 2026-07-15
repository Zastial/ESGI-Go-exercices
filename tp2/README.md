# Mira — API v1 (`/api/v1/notes`)

API Notes du projet fil rouge Mira. Stockage en mémoire (`map` +
`sync.RWMutex`) pour l'instant.

## Structure

```
cmd/api/main.go                point d'entrée: logger, store, routeur, serveur, shutdown (+ annotations @title/@description)
cmd/api/docs/                   spec OpenAPI générée par swag init (docs.go, swagger.json, swagger.yaml)
internal/core/note.go           domaine: Note, CreateNoteInput/UpdateNoteInput, NoteStore (interface), erreurs
internal/store/memory.go        implémentation en mémoire de core.NoteStore
internal/http/router.go         assemble routes + middlewares + /swagger/ (NewRouter)
internal/http/handlers/notes.go handlers HTTP (POST/GET/PATCH/DELETE/search) + helpers + annotations swag
internal/http/middleware/       request ID, logging (slog), recovery, timeout
internal/http/response/         enveloppe JSON stable (Envelope / JSON / Error)
```

Le routeur est séparé de `main.go` : `internal/http/router.go` expose
`NewRouter(store, logger) http.Handler`, qui déclare les routes et empile les
middlewares. `main.go` ne fait plus que l'appeler et démarrer le serveur.

Stockage : `MemoryStore` génère les IDs avec `github.com/google/uuid`, garde
`CreatedAt` et applique uniquement les champs fournis lors d'un `PATCH`
(`UpdateNoteInput` utilise des pointeurs pour distinguer "non fourni" de
"fourni vide"). `Search` fait un `strings.Contains` insensible à la casse sur
`Title`/`Content`.

## Lancer le serveur

```bash
go run ./cmd/api
# écoute sur :8080 (ou $PORT)
```

## Lancer les tests

```bash
go test ./...
```

`internal/http/handlers/notes_test.go` couvre : création (succès + 400 sur
payload invalide) et récupération d'une note inexistante (404).

## Endpoints

| Méthode | Route                  | Description                  |
|---------|-------------------------|-------------------------------|
| POST    | `/api/v1/notes`         | Créer une note                |
| GET     | `/api/v1/notes`         | Lister les notes (paginé)     |
| GET     | `/api/v1/notes/{id}`    | Récupérer une note             |
| PATCH   | `/api/v1/notes/{id}`    | Mettre à jour partiellement    |
| DELETE  | `/api/v1/notes/{id}`    | Supprimer une note             |
| GET     | `/api/v1/search?q=...`  | Recherche texte simple (paginée)|

### Pagination (bonus)

`GET /api/v1/notes` et `GET /api/v1/search` acceptent `?limit=&offset=`
(défaut `limit=20`, plafonné à `100`; `offset` défaut `0`). Le total (avant
pagination) est renvoyé dans `meta.total`.

## Enveloppe de réponse

Toutes les réponses suivent la même forme stable :

```json
{
  "data": { ... },
  "error": null,
  "meta": null
}
```

- `data` : présent sur succès (objet ou tableau selon l'endpoint)
- `error` : `{ "code": "...", "message": "..." }` présent sur erreur
- `meta` : `{ "limit": 20, "offset": 0, "total": 42 }` sur les endpoints paginés

## Exemples `curl`

```bash
# Créer une note
curl -i -X POST http://localhost:8080/api/v1/notes \
  -H 'Content-Type: application/json' \
  -d '{"title":"Courses","content":"Lait, œufs, pain"}'

# Lister (paginé)
curl -i "http://localhost:8080/api/v1/notes?limit=10&offset=0"

# Récupérer une note
curl -i http://localhost:8080/api/v1/notes/<id>

# Mettre à jour partiellement
curl -i -X PATCH http://localhost:8080/api/v1/notes/<id> \
  -H 'Content-Type: application/json' \
  -d '{"content":"Lait, œufs, pain, beurre"}'

# Supprimer
curl -i -X DELETE http://localhost:8080/api/v1/notes/<id>

# Recherche
curl -i "http://localhost:8080/api/v1/search?q=courses&limit=10&offset=0"
```

## Codes d'erreur possibles

| Status | Code               | Quand                                              |
|--------|--------------------|-----------------------------------------------------|
| 400    | `VALIDATION_ERROR` | payload invalide (JSON malformé ou règles métier)    |
| 404    | `NOT_FOUND`        | note inexistante (GET/PATCH/DELETE par id)           |
| 408/504| `TIMEOUT`          | la requête dépasse le timeout serveur (5s par défaut)|
| 500    | `INTERNAL_ERROR`   | erreur interne / panique récupérée par Recovery      |

Le mapping erreur métier → status HTTP est déjà centralisé dans
`internal/http/handlers/notes.go` (`writeError`) : il reconnaît
`core.ErrNotFound` (→ 404) et `core.ErrValidation` (→ 400, via `errors.Is`
sur une erreur wrappée avec `%w`), et retombe sur 500 sinon.

## Middlewares (déjà branchés dans `internal/http/router.go`)

Ordre d'exécution : `RequestID` → `Logging` → `Recovery` → `Timeout` → routes.

- **RequestID** : génère/relaie un `X-Request-ID`, disponible via
  `middleware.RequestIDFromContext(ctx)`.
- **Logging** : une ligne `slog` structurée par requête (method, path, status,
  durée, request_id).
- **Recovery** : capture les panics et répond `500` proprement.
- **Timeout** : annule la requête et répond `408`-style JSON après 5s.

## OpenAPI / Swagger (bonus)

La spec est générée avec [`swaggo/swag`](https://github.com/swaggo/swag) à
partir d'annotations sur `cmd/api/main.go` (infos générales) et sur chaque
handler de `internal/http/handlers/notes.go`. Le résultat vit dans
`cmd/api/docs/` (`docs.go`, `swagger.json`, `swagger.yaml`) et est servi par
[`swaggo/http-swagger`](https://github.com/swaggo/http-swagger), branché dans
`internal/http/router.go`.

```bash
# UI Swagger, une fois le serveur lancé (go run ./cmd/api)
open http://localhost:8080/swagger/index.html

# régénérer la doc après avoir modifié des annotations @... ou des types
# core.Note / core.CreateNoteInput / core.UpdateNoteInput / response.ErrorBody
swag init -g cmd/api/main.go -o cmd/api/docs --parseDependency --parseInternal
```

`cmd/api/docs/docs.go` est généré (`DO NOT EDIT`) — ne pas le modifier à la
main, relancer `swag init` à la place.
