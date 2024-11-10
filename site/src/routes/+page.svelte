<script lang="ts">
    import Board, { type Cell } from "$lib/Board.svelte";
    import Header from "$lib/Header.svelte";

    import { PUBLIC_API_URL } from '$env/static/public';
    import {onMount} from "svelte";

    let standardModifiers = $state<Cell[]>([]);

    onMount(async () => {
        const response = await fetch(`${PUBLIC_API_URL}/api/v1/presets/standard/board`);
        standardModifiers = await response.json().then(data => data.cells);
    });
</script>

<Header />
<main class="container">
    <section class="container">
        <h2>Get started</h2>
        <p>
            To get started, click "Create a game", enter your name, then send the
            link to your friends or enemies!
        </p>
        <a href="/new" class="button">Create a game</a>
    </section>
    <section class="container">
        <h2>How to play</h2>
        <p>
            This word game challenges players to create words on a grid-based
            board, scoring points based on letter values and special board spaces.
        </p>
        <h3>GameSvelte Setup</h3>
        <ul>
            <li>The game is played on a grid.</li>
            <li>Each player is given tiles to begin.</li>
            <li>Players take turns placing words on the board.</li>
        </ul>
        <h3>Gameplay</h3>
        <ol>
            <li>The first word must cover the center space.</li>
            <li>
                Subsequent words must connect to or intersect with existing words on
                the board.
            </li>
            <li>
                Words can be placed horizontally (left to right) or vertically (top
                to bottom).
            </li>
            <li>
                All words formed during a turn must be deemed valid by the other
                players.
            </li>
            <li>
                After playing a word, new tiles will be given to replace those used.
            </li>
        </ol>
        <h3>Scoring</h3>
        <ul>
            <li>Each letter has a point value.</li>
            <li>
                Special board spaces can modify letter or word scores:
                <Board
                        offsetX={-15}
                        offsetY={-15}
                        width={4*60+30}
                        height={60+30}
                        cellSize={60}
                        cells={[
                            { x: 0, y: 0, modifier: "DL" },
                            { x: 1, y: 0, modifier: "TL" },
                            { x: 2, y: 0, modifier: "DW" },
                            { x: 3, y: 0, modifier: "TW" },
                        ]}
                        disabled
                />
                <ul>
                    <li>
                        Double Letter Score (DL)
                    </li>
                    <li>Triple Letter Score (TL)</li>
                    <li>Double Word Score (DW)</li>
                    <li>Triple Word Score (TW)</li>
                </ul>
            </li>
            <li>Modifiers apply only for the first letter placed on them.</li>
            <li>Score all words formed or modified during your turn.</li>
            <li>
                The main word score is calculated first, then any additional words
                formed.
            </li>
        </ul>
        <h3>Indirect Word Formation</h3>
        <p>
            When you place a word, you may form additional words perpendicular to
            your main word. These "indirect" words are also scored and must be
            valid. For example:
        </p>
        <Board
                offsetX={-30}
                offsetY={-30}
                width={5 * 60}
                height={3 * 60}
                cellSize={60}
                cells={[
                    { x: 0, y: 0, letter: "H" },
                    { x: 1, y: 0, letter: "E" },
                    { x: 2, y: 0, letter: "Y" },
                    { x: 2, y: 1, letter: "O" },
                    { x: 3, y: 1, letter: "N" },
                    ...standardModifiers,
                ]}
                    letterPoints={{
                    "H": 4,
                    "E": 1,
                    "Y": 4,
                    "O": 1,
                    "N": 1,
                }}
                disabled
        />
        <p>
            In this example, playing "ON" horizontally not only scores for "ON"
            but also creates and scores the vertical word "YO".
        </p>
        <h3>Ending the GameSvelte</h3>
        <p>
            The game ends when:
        </p>
        <ul>
            <li>
                All tiles have been drawn and one player uses all their tiles, or
            </li>
            <li>No more valid plays can be made</li>
        </ul>

        <h3>Winning</h3>
        <p>
            The player with the highest total score at the end of the game wins.
        </p>
    </section>
</main>