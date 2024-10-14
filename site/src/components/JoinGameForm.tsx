import { useState } from "react";

type Props = {
  gameId: string;
  joinGame: (gameId: string, playerId: string) => void;
};

export default function JoinGame({ gameId, joinGame }: Props) {
  const [name, setName] = useState("");
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        joinGame(gameId, name);
      }}
    >
      <h3>Join game {gameId}</h3>
      <label>
        Game ID
        <input type="text" value={gameId} disabled />
      </label>
      <label>
        Player Name
        <input
          type="text"
          name={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
      </label>
      <button type="submit">Join</button>
    </form>
  );
}
