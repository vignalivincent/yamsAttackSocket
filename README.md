# YamsAttackSocket

Un serveur simple pour le jeu Yams Attack, écrit en Go.

## Prérequis

- Go 1.21 ou supérieur
- Docker (optionnel)
- Fly.io CLI (pour le déploiement)

## Installation et exécution

### Exécution locale

1. Clonez le dépôt
2. Exécutez le serveur :

```bash
go run main.go
```

Le serveur démarre sur le port 8080.

### Utilisation de Docker

1. Construisez l'image Docker :

```bash
docker build -t yams-attack-socket .
```

2. Exécutez le conteneur :

```bash
docker run -p 8080:8080 yams-attack-socket
```

## Test du serveur

Utilisez ces commandes curl pour tester l'API :

```bash
# Vérifier la santé du serveur
curl -X GET http://localhost:8080/health

# Obtenir l'état du jeu
curl -X GET http://localhost:8080/game

# Ajouter un joueur
curl -X POST http://localhost:8080/player \
  -H "Content-Type: application/json" \
  -d '{"id":"1", "name":"Joueur 1", "score":0}'

# Vérifier l'état du jeu après ajout
curl -X GET http://localhost:8080/game
```

## Déploiement sur Fly.io

### 1. Installation du CLI Fly.io

Si ce n'est pas déjà fait, installez le CLI Fly.io :

```bash
curl -L https://fly.io/install.sh | sh
```

Ou sur macOS avec Homebrew :

```bash
brew install flyctl
```

### 2. Authentification

Connectez-vous à votre compte Fly.io :

```bash
fly auth login
```

### 3. Déploiement

Lancez le déploiement :

```bash
# Si c'est votre première fois
fly launch

# Pour les déploiements suivants
fly deploy
```

L'application sera accessible à l'adresse `https://yams-attack-socket.fly.dev`

### 4. Surveillance et logs

Pour voir les logs de l'application :

```bash
fly logs
```

## API Endpoints

### GET /health

Vérifie que le serveur fonctionne correctement.

### GET /game

Récupère l'état actuel du jeu.

### POST /player

Ajoute un nouveau joueur au jeu.

Exemple de requête :

```json
{
  "id": "1",
  "name": "Player 1",
  "score": 0
}
```
