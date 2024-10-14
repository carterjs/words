import { useState } from "react";

import styles from "./Forms.module.css";

type Props = {
  createGame: (playerName: string) => void;
};

export default function CreateGame({ createGame }: Props) {
  const [name, setName] = useState("");

  return (
    <form
      className={styles.form}
      onSubmit={(e) => {
        e.preventDefault();
        createGame(name);
      }}
    >
      <h2 className={styles.title}>Create a Game</h2>
      <label className={styles.field}>
        Player/Team Name
        <input
          type="text"
          placeholder="Your Name"
          name={name}
          onChange={(e) => setName(e.target.value)}
          required
          autoFocus
        />
      </label>
      <button type="submit" className="button">
        Create Game
      </button>
    </form>
  );
}
