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

## 🔌 Guide d'utilisation des API

### Flux d'utilisation typique

1. **Joueur principal**: Joue au Yams sur l'application frontend
2. **Partage**: Le joueur décide de partager sa partie via l'API REST `/share`
3. **SMS**: Les spectateurs reçoivent un lien par SMS
4. **Connexion WebSocket**:
   - Le joueur principal se connecte pour envoyer ses mises à jour
   - Les spectateurs se connectent pour recevoir les mises à jour
5. **Temps réel**: Le joueur continue sa partie, les spectateurs voient l'évolution en direct

### API REST détaillée - /share

#### Endpoint: POST `/share`

**Description**: Initie le partage d'une partie avec des spectateurs via SMS. Le serveur générera un identifiant unique, créera une URL de visualisation et enverra des SMS aux numéros indiqués.

**Headers**:

```
Content-Type: application/json
```

**Request Body**:

```json
{
  "playerId": "player123", // ID unique du joueur principal (obligatoire)
  "gameData": {
    // État actuel du jeu (obligatoire)
    "players": [
      {
        "id": "player123",
        "name": "Joueur Principal",
        "scores": {
          "ones": 3,
          "twos": 6
          // autres scores...
        }
      }
      // autres joueurs...
    ],
    "isStarted": true,
    "gameHistory": [] // historique optionnel
  },
  "phoneNumbers": ["33612345678"] // Numéros des spectateurs (min 1)
}
```

**Response**:

```json
{
  "success": true,
  "gameId": "550e8400-e29b-41d4-a716-446655440000",
  "viewUrl": "http://votre-domaine.com?view=true&gameId=550e8400-e29b-41d4-a716-446655440000"
}
```

**Actions automatiques**:

- Génération d'un ID unique pour la partie
- Envoi de SMS avec le lien aux spectateurs
- Stockage de l'état initial de la partie

## 📊 Formats de données et codes de réponse

### API REST: POST `/share`

#### Structure des données

**Requête:**

```typescript
interface ShareRequest {
  // ID unique du joueur principal (obligatoire)
  playerId: string;

  // État actuel du jeu (obligatoire)
  gameData: {
    // Liste des joueurs dans la partie
    players: Array<{
      id: string; // ID unique du joueur
      name: string; // Nom d'affichage du joueur
      scores: {
        // Scores du joueur (tous optionnels)
        ones?: number | 'crossed';
        twos?: number | 'crossed';
        threes?: number | 'crossed';
        fours?: number | 'crossed';
        fives?: number | 'crossed';
        sixes?: number | 'crossed';
        threeOfAKind?: number | 'crossed';
        fourOfAKind?: number | 'crossed';
        fullHouse?: number | 'crossed';
        smallStraight?: number | 'crossed';
        largeStraight?: number | 'crossed';
        yahtzee?: number | 'crossed';
        chance?: number | 'crossed';
      };
    }>;

    // Indique si la partie a commencé
    isStarted: boolean;

    // Historique des parties précédentes (optionnel)
    gameHistory?: Array<{
      id: string; // ID de la partie historique
      date: string; // Date au format ISO (ex: "2025-03-15T12:31:29.304Z")
      players: Array<{
        id: string; // ID du joueur
        name: string; // Nom du joueur
        score: number; // Score final
      }>;
      winnerId: string; // ID du joueur gagnant
    }>;
  };

  // Numéros de téléphone des spectateurs (minimum 1)
  phoneNumbers: string[];
}
```

**Réponse:**

```typescript
interface ShareResponse {
  // Indique si la requête a réussi
  success: boolean;

  // ID unique généré pour la partie partagée
  gameId: string;

  // URL à partager avec les spectateurs
  viewUrl: string;

  // Message d'erreur (uniquement en cas d'échec)
  error?: string;
}
```

#### Codes de statut

| Code | Description           | Situation                                                                  |
| ---- | --------------------- | -------------------------------------------------------------------------- |
| 200  | OK                    | Partage réussi                                                             |
| 400  | Bad Request           | Format JSON invalide, `playerId` manquant, ou liste de `phoneNumbers` vide |
| 405  | Method Not Allowed    | Méthode autre que POST utilisée                                            |
| 500  | Internal Server Error | Erreur serveur inattendue                                                  |

### API WebSocket: `/ws`

#### Structure des messages

**Messages client → serveur:**

```typescript
interface ClientMessage {
  // Type de message
  type: 'join_game' | 'update_game_state' | 'leave_game' | 'view_request';

  // ID de la partie
  gameId: string;

  // Contenu spécifique au type de message
  content: any;
}
```

**Exemples par type:**

1. **join_game** - Rejoindre une partie

```typescript
{
  type: "join_game",
  gameId: "550e8400-e29b-41d4-a716-446655440000",
  content: {
    playerId: "player123",
    playerName: "Joueur 1"
  }
}
```

2. **update_game_state** - Mettre à jour l'état du jeu

```typescript
{
  type: "update_game_state",
  gameId: "550e8400-e29b-41d4-a716-446655440000",
  content: {
    state: {
      // Structure identique à gameData dans ShareRequest
      players: [...],
      isStarted: true,
      gameHistory: [...]
    },
    version: 2  // Numéro de version (incrémental)
  }
}
```

3. **leave_game** - Quitter une partie

```typescript
{
  type: "leave_game",
  gameId: "550e8400-e29b-41d4-a716-446655440000",
  content: {}
}
```

**Messages serveur → client:**

```typescript
interface ServerMessage {
  // Type de message
  type: 'game_state_updated' | 'player_joined' | 'error';

  // ID de la partie
  gameId: string;

  // Contenu spécifique au type de message
  content: any;
}
```

#### Codes d'erreur WebSocket

| Code | Description    | Situation                                      |
| ---- | -------------- | ---------------------------------------------- |
| 1000 | Normal Closure | Fermeture normale de la connexion              |
| 1001 | Going Away     | Le client se déconnecte ou le serveur s'arrête |
| 1011 | Internal Error | Erreur interne du serveur                      |

## 💻 Exemple d'implémentation JavaScript

### Connexion au WebSocket (`/ws`)

**Description**: Permet aux joueurs et aux spectateurs de se connecter pour envoyer et recevoir des mises à jour en temps réel.

**Headers**:

```
Sec-WebSocket-Protocol: chat
```

**Request**:

```
GET /ws?gameId=550e8400-e29b-41d4-a716-446655440000
```

**Messages**:

- **Joueur principal**: Envoie des mises à jour de l'état du jeu
- **Spectateurs**: Reçoivent les mises à jour en temps réel

**Exemple de message envoyé par le joueur principal**:

```json
{
  "type": "update",
  "gameData": {
    "players": [
      {
        "id": "player123",
        "name": "Joueur Principal",
        "scores": {
          "ones": 3,
          "twos": 6
          // autres scores...
        }
      }
      // autres joueurs...
    ],
    "isStarted": true,
    "gameHistory": [] // historique optionnel
  }
}
```

**Exemple de message reçu par les spectateurs**:

```json
{
  "type": "update",
  "gameData": {
    "players": [
      {
        "id": "player123",
        "name": "Joueur Principal",
        "scores": {
          "ones": 3,
          "twos": 6
          // autres scores...
        }
      }
      // autres joueurs...
    ],
    "isStarted": true,
    "gameHistory": [] // historique optionnel
  }
}
```
