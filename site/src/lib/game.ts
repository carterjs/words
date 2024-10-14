import { useEffect, useReducer, type Reducer } from "react";
import useWebSocket from "react-use-websocket";

type Message = {
  type: MessageType;
  payload: any;
};

enum MessageType {
  JoinGame = "join_game",
  RejoinGame = "rejoin_game",
  CreateGame = "create_game",
  StartGame = "start_game",
}

export type GameState = null | {
  gameId: string;
  playerId: string;
  started: boolean;
  players: Player[];
  letterPoints: { [letter: string]: number };
  rack: string[];
  turn: number;
  grid: { [point: string]: string };
};

export type Player = {
  id: string;
  name: string;
};

export function useGame(url: string) {
  // initialize game state reducer
  const [state, dispatch] = useReducer<Reducer<GameState, Message>>(
    serverResponseReducer,
    null
  );

  // connect to websocket server
  const { sendJsonMessage, lastMessage, readyState } = useWebSocket(url, {
    retryOnError: true,
  });

  // TODO: add connection state into game state
  useEffect(() => {}, [readyState]);

  // pipe messages through reduce
  useEffect(() => {
    if (!lastMessage) {
      return;
    }

    const message = JSON.parse(lastMessage.data);
    dispatch(message);
  }, [lastMessage]);

  // return game state and actions
  return {
    game: state,

    // createGame creates a new game and adds the player to it
    createGame(playerName: string) {
      sendJsonMessage({
        type: MessageType.CreateGame,
        payload: {
          playerName,
        },
      });
    },

    // joinGame adds a player to an existing game and will error if the game doesn't exist
    joinGame(gameId: string, playerName: string) {
      sendJsonMessage({
        type: MessageType.JoinGame,
        payload: {
          gameId,
          playerName,
        },
      });
    },

    // rejoin game will connect a player to the game they're already in
    rejoinGame(gameId: string, playerId: string) {
      sendJsonMessage({
        type: MessageType.RejoinGame,
        payload: {
          gameId,
          playerId,
        },
      });
    },

    startGame() {
      sendJsonMessage({
        type: MessageType.StartGame,
        payload: {},
      });
    },
  };
}

// reducer to read events from the server and update the game state accordingly
const serverResponseReducer = (
  state: GameState,
  message: Message
): GameState => {
  console.log("processing message", message);

  switch (message.type) {
    case MessageType.CreateGame:
      return {
        gameId: message.payload.gameId,
        playerId: message.payload.playerId,
        started: false,
        players: message.payload.players,
        letterPoints: message.payload.letterPoints,
        rack: [],
        turn: 0,
        grid: message.payload.grid,
      };
    case MessageType.JoinGame:
      return {
        gameId: message.payload.gameId,
        playerId: message.payload.playerId,
        started: message.payload.started,
        players: message.payload.players,
        letterPoints: message.payload.letterPoints,
        // rack empty since the game hasn't started
        rack: [],
        turn: 0,
        grid: message.payload.grid,
      };
    case MessageType.RejoinGame:
      return {
        gameId: message.payload.gameId,
        playerId: message.payload.playerId,
        started: message.payload.started,
        players: message.payload.players,
        letterPoints: message.payload.letterPoints,
        rack: message.payload.rack,
        turn: message.payload.turn,
        grid: message.payload.grid,
      };
    case MessageType.StartGame:
      if (!state) {
        return state;
      }

      return {
        ...state,
        rack: message.payload.rack,
        started: true,
      };
  }

  return state;
};
