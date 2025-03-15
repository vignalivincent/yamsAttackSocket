# Guide de l'API Yams Attack pour développeurs frontend

## Introduction

Ce guide explique comment utiliser l'API Yams Attack dans votre application frontend. Il y a deux façons d'interagir avec l'API:

1. **API REST** - Pour partager une partie avec des spectateurs
2. **WebSocket** - Pour les communications en temps réel pendant la partie

## 1. Partager une partie (API REST)

### Endpoint

`POST /share`

### Exemple avec fetch

```javascript
// Fonction pour partager une partie
async function partagerPartie(playerId, gameData, phoneNumbers) {
  try {
    const response = await fetch('http://localhost:8080/share', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        playerId: playerId,
        gameData: gameData,
        phoneNumbers: phoneNumbers,
      }),
    });

    if (!response.ok) throw new Error('Erreur lors du partage');

    const result = await response.json();

    console.log('Partie partagée!');
    console.log('ID de la partie:', result.gameId);
    console.log('URL pour les spectateurs:', result.viewUrl);

    return result;
  } catch (error) {
    console.error('Erreur:', error);
    return null;
  }
}

// Exemple d'utilisation
const etatJeu = {
  players: [
    {
      id: 'player123',
      name: 'Joueur 1',
      scores: {
        ones: 3,
        twos: 6,
      },
    },
  ],
  isStarted: true,
  gameHistory: [],
};

// Appel de la fonction
partagerPartie('player123', etatJeu, ['33612345678']);
```

## 2. Communication en temps réel (WebSocket)

### Connexion WebSocket

#### Pour un joueur

```javascript
// Créer une connexion WebSocket
const socket = new WebSocket('ws://localhost:8080/ws');

socket.onopen = function () {
  console.log('Connecté au serveur WebSocket');

  // Rejoindre une partie après connexion
  rejoindrePartie('game123', 'player1', 'Joueur 1');
};

socket.onmessage = function (event) {
  const message = JSON.parse(event.data);
  console.log('Message reçu:', message);

  // Traiter le message selon son type
  switch (message.type) {
    case 'game_state_updated':
      mettreAJourInterface(message.content.state);
      break;
    case 'player_joined':
      afficherNotification(`${message.content.playerName} a rejoint la partie!`);
      break;
  }
};

socket.onclose = function () {
  console.log('Connexion fermée');
};
```

#### Pour un spectateur

```javascript
// URL avec paramètres pour spectateur
const url = new URL('ws://localhost:8080/ws');
url.searchParams.append('view', 'true');
url.searchParams.append('gameId', 'GAME_ID_REÇU');

const socket = new WebSocket(url.toString());

socket.onmessage = function (event) {
  const message = JSON.parse(event.data);
  if (message.type === 'game_state_updated') {
    // Mettre à jour l'interface avec le nouvel état
    afficherEtatJeu(message.content.state);
  }
};
```

## 3. Messages à envoyer (joueur uniquement)

### Rejoindre une partie

```javascript
function rejoindrePartie(gameId, playerId, playerName) {
  const message = {
    type: 'join_game',
    gameId: gameId,
    content: {
      playerId: playerId,
      playerName: playerName,
    },
  };

  socket.send(JSON.stringify(message));
}
```

### Mettre à jour l'état du jeu

```javascript
function envoyerMiseAJour(gameId, etatJeu, version) {
  const message = {
    type: 'update_game_state',
    gameId: gameId,
    content: {
      state: etatJeu,
      version: version,
    },
  };

  socket.send(JSON.stringify(message));
}
```

### Quitter une partie

```javascript
function quitterPartie(gameId) {
  const message = {
    type: 'leave_game',
    gameId: gameId,
    content: {},
  };

  socket.send(JSON.stringify(message));
}
```

## 4. Messages reçus

### État du jeu mis à jour

```javascript
// Dans votre gestionnaire de message (socket.onmessage)
if (message.type === 'game_state_updated') {
  const etatJeu = message.content.state;
  const version = message.content.version;

  // Exemple: afficher les joueurs et leurs scores
  etatJeu.players.forEach((player) => {
    console.log(`Joueur: ${player.name}`);
    console.log('Scores:', player.scores);
  });
}
```

### Nouveau joueur rejoint

```javascript
if (message.type === 'player_joined') {
  const newPlayerId = message.content.playerId;
  const newPlayerName = message.content.playerName;

  console.log(`${newPlayerName} a rejoint la partie!`);
}
```

## 5. Exemple complet: Flux de jeu typique

```javascript
// 1. Connexion WebSocket
const socket = new WebSocket('ws://localhost:8080/ws');

// Variables de jeu
const gameId = 'game123';
const playerId = 'player' + Math.floor(Math.random() * 1000);
const playerName = 'Joueur ' + Math.floor(Math.random() * 100);
let gameVersion = 1;
let gameState = {
  players: [],
  isStarted: true,
  gameHistory: [],
};

// 2. Quand la connexion est établie
socket.onopen = function () {
  // Rejoindre une partie existante ou après partage
  rejoindrePartie(gameId, playerId, playerName);
};

// 3. Réception des messages
socket.onmessage = function (event) {
  const message = JSON.parse(event.data);

  switch (message.type) {
    case 'game_state_updated':
      // Mettre à jour l'état local
      gameState = message.content.state;
      gameVersion = message.content.version;
      rafraichirInterface();
      break;

    case 'player_joined':
      notifierNouveauJoueur(message.content.playerName);
      break;
  }
};

// 4. Fonctions utilitaires
function rejoindrePartie(gameId, playerId, playerName) {
  socket.send(
    JSON.stringify({
      type: 'join_game',
      gameId: gameId,
      content: { playerId, playerName },
    })
  );
}

function mettreAJourScore(combinaison, valeur) {
  // Trouver le joueur actuel
  const monJoueur = gameState.players.find((p) => p.id === playerId);
  if (!monJoueur) return;

  // Mettre à jour le score
  if (!monJoueur.scores) monJoueur.scores = {};
  monJoueur.scores[combinaison] = valeur;

  // Incrémenter la version
  gameVersion++;

  // Envoyer l'état mis à jour
  socket.send(
    JSON.stringify({
      type: 'update_game_state',
      gameId: gameId,
      content: {
        state: gameState,
        version: gameVersion,
      },
    })
  );
}

// Exemple: déclenché par l'UI quand un joueur inscrit un score
document.getElementById('inscrireScore').addEventListener('click', function () {
  const combinaison = document.getElementById('combinaison').value; // ex: "ones"
  const valeur = parseInt(document.getElementById('valeur').value);
  mettreAJourScore(combinaison, valeur);
});
```

## Format des données

### État du jeu (gameData / state)

```javascript
{
  "players": [
    {
      "id": "player123",      // ID unique du joueur
      "name": "Joueur 1",     // Nom du joueur
      "scores": {             // Scores (propriétés facultatives)
        "ones": 3,            // As
        "twos": 6,            // Deux
        "threes": 9,          // Trois
        "fours": 12,          // Quatre
        "fives": 15,          // Cinq
        "sixes": 24,          // Six
        "threeOfAKind": 30,   // Brelan
        "fourOfAKind": 40,    // Carré
        "fullHouse": 25,      // Full
        "smallStraight": 30,  // Petite suite
        "largeStraight": 40,  // Grande suite
        "yahtzee": 50,        // Yams
        "chance": 20          // Chance
      }
    }
  ],
  "isStarted": true,          // Indique si la partie a commencé
  "gameHistory": []           // Historique (optionnel)
}
```
