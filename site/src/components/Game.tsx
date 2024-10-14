import JoinGameForm from "./JoinGameForm";
import CreateGameForm from "./CreateGameForm";
import { useGame } from "../lib/game";
import { useEffect, useState } from "react";
import Board from "./Board";

import styles from "./Game.module.css";

type Props = {
  websocketUrl: string;
};

let initialGameId = new URLSearchParams(window.location.search).get("game");
let initialPlayerId = localStorage.getItem(initialGameId + "/playerId");

export default function Block({ websocketUrl = "/ws" }: Props) {
  const [gameId, setGameId] = useState(initialGameId);
  const [playerId, setPlayerId] = useState(initialPlayerId);
  const [boardWidth, setBoardWidth] = useState(window.innerWidth);
  const [boardHeight, setBoardHeight] = useState(window.innerHeight);

  useEffect(() => {
    const updateSize = () => {
      setBoardWidth(window.innerWidth);
      setBoardHeight(window.innerHeight);
    };

    window.addEventListener("resize", updateSize);
    return () => window.removeEventListener("resize", updateSize);
  }, []);

  const { game, createGame, joinGame, rejoinGame, startGame } =
    useGame(websocketUrl);

  // join the game if we have all we need
  useEffect(() => {
    if (gameId && playerId && !game) {
      rejoinGame(gameId, playerId);
    }
  }, [gameId, playerId]);

  useEffect(() => {
    if (!game) {
      return;
    }

    setGameId(game.gameId);
    setPlayerId(game.playerId);
    window.history.pushState({}, "", `?game=${game.gameId}`);
    localStorage.setItem(game.gameId + "/playerId", game.playerId);
  }, [game]);

  if (!game) {
    return (
      <div className={styles.form}>
        {!gameId && <CreateGameForm createGame={createGame} />}
        {gameId && !playerId && (
          <JoinGameForm gameId={gameId} joinGame={joinGame} />
        )}
      </div>
    );
  }

  return (
    <>
      <h1>Game: {game.gameId}</h1>
      {/* <table>
        <thead>
          <tr>
            <th>Letter</th>
            <th>Point Value</th>
          </tr>
        </thead>
        <tbody>
          {Object.keys(game.letterPoints).map((letter) => (
            <tr key={letter}>
              <td>{letter}</td>
              <td>{game.letterPoints[letter]}</td>
            </tr>
          ))}
        </tbody>
      </table> */}
      {/* <ul>
        {game.players.map((player) => (
          <li key={player.id}>{player.name}</li>
        ))}
      </ul> */}
      {game.grid && (
        <Board
          grid={game.grid}
          width={boardWidth}
          height={boardHeight}
          offsetX={-boardWidth / 2 + 25}
          offsetY={-boardHeight / 2 + 25}
          cellSize={50}
          fullScreen
          allowPanning
        />
      )}
      <div className={styles.controls}>
        <div className="container">
          {!game.started && (
            <button onClick={startGame} className="button">
              Start Game
            </button>
          )}
          <ul className={styles.rack}>
            {game.rack &&
              game.rack.map((letter, i) => <li key={i}>{letter}</li>)}
          </ul>
        </div>
      </div>
    </>
  );
}
