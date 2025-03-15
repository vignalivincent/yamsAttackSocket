# Yams Attack Socket

API WebSocket et REST pour le partage et la visualisation en temps réel des parties du jeu Yams Attack.

## 📋 À propos

Yams Attack Socket est une API permettant aux joueurs de:

- Partager des parties en cours via SMS avec des spectateurs
- Communiquer en temps réel via WebSocket pour synchroniser l'état du jeu
- Permettre aux spectateurs de suivre l'évolution des parties sans pouvoir intervenir

## 🚀 Installation et démarrage

### Prérequis

- Go 1.16+
- Connexion internet pour les dépendances

### Installation

```bash
# Cloner le dépôt
git clone https://github.com/vignaliVincent/yamsAttackSocket.git
cd yamsAttackSocket

# Installer les dépendances
go mod download
```

### Démarrage du serveur

```bash
# Compiler et exécuter
go run cmd/server/main.go

# Ou construire et exécuter
go build -o yamsAttack cmd/server/main.go
./yamsAttack
```

Le serveur démarre par défaut sur le port 8080. Pour changer le port:

```bash
PORT=9000 go run cmd/server/main.go
```

## 🔌 APIs disponibles

### API REST

- **POST `/share`**: Partager une partie avec des spectateurs via SMS

### API WebSocket

- **WebSocket `/ws`**: Connexion pour joueurs et spectateurs

## 📱 Pages de test

Des interfaces HTML sont disponibles pour tester l'API:

- **Test Joueur**: <http://localhost:8080/test/player_test.html>
- **Test Spectateur**: <http://localhost:8080/test/viewer_test.html?view=true&gameId=VOTRE_GAME_ID>

## 📖 Documentation

Une documentation détaillée est disponible dans le dossier `docs/`:

- [Guide de l'API pour développeurs frontend](docs/api_reference.md)

## 📐 Architecture du projet
