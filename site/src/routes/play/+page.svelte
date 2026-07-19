<script lang="ts">
    import {GameController} from '$lib/game.svelte';
    import { page } from '$app/stores';
    import {onMount} from "svelte";
    import Board from "$lib/Board.svelte";
    import Rack from "$lib/Rack.svelte";

    let gameId = $state<string>("");
    let game = new GameController();

    let boardWidth = $state(0);
    let boardHeight = $state(0);
    let offsetX = $state(0);
    let offsetY = $state(0);

    // when the component mounts, get game info using the search param
    onMount(async () => {
        gameId = $page.url.searchParams.get("game") || "";
        if (!gameId) {
            return
        }

        game.id = gameId;

        await game.loadInitialData();
        game.streamUpdates()

        applyWindowSize();

        // start with 0,0 centered
        offsetY = -boardHeight / 2 + 30;
        offsetX = -boardWidth / 2 + 30;
    });

    // Set the board resolution to screen size
    function applyWindowSize() {
        boardWidth = window.innerWidth;
        boardHeight = window.innerHeight;
    }

    let name = $state("");

    $effect(() => {
        game.input = game.input.toUpperCase().replace(/[^A-Z]/g, "");
    })

    let selectedCell = $state<{ x: number; y: number } | null>(null);

    async function handleCellTap(x: number, y: number) {
        if (!game.started || game.finished || game.challenge) {
            return;
        }

        if (!game.myTurn) {
            game.error = `It's ${game.playerName(game.currentPlayerId)}'s turn.`;
            return;
        }

        if (!game.input) {
            game.error = "Type a word first, then tap where it goes.";
            return;
        }

        selectedCell = { x, y };
        await game.findPlacements(x, y);

        if (game.placements.length === 0 && !game.error) {
            game.error = "That word doesn't fit there.";
            selectedCell = null;
        }
    }

    function cancelPlacement() {
        selectedCell = null;
        game.clearPlacements();
    }

    async function playWord() {
        await game.playSelectedPlacement();
        if (!game.error) {
            selectedCell = null;
        }
    }

    // ghost tiles previewing the currently selected placement option;
    // letters already on the board are left out
    let ghostCells = $derived.by(() => {
        const placement = game.placements[game.selectedPlacement];
        if (!placement) return [];

        const ghosts = [];
        for (let i = 0; i < placement.word.length; i++) {
            const x = placement.direction === "HORIZONTAL" ? placement.x + i : placement.x;
            const y = placement.direction === "VERTICAL" ? placement.y + i : placement.y;

            const occupied = game.board.cells.some(cell => cell.x === x && cell.y === y && cell.letter);
            if (!occupied) {
                ghosts.push({ x, y, letter: placement.word[i] });
            }
        }

        return ghosts;
    });

    let selectedPlacement = $derived(game.placements[game.selectedPlacement]);

    let canChallenge = $derived(
        game.started
        && !game.finished
        && !game.challenge
        && game.challengeableMoverId !== null
        && game.challengeableMoverId !== game.playerId
    );

    let boardComponent = $state<ReturnType<typeof Board> | undefined>();

    let boardView = $state<{ minX: number; minY: number; maxX: number; maxY: number } | null>(null);

    // offer a way home when the center star has drifted off-screen
    let showRecenter = $derived(
        boardView !== null
        && (boardView.maxX < 0 || boardView.minX > 1 || boardView.maxY < 0 || boardView.minY > 1)
    );

    // cells of the most recently played word, tinted on the board
    let lastWordCells = $derived.by(() => {
        const lastWord = game.lastWord;
        if (!lastWord) return [];

        return [...lastWord.word].map((_, i) => ({
            x: lastWord.direction === "HORIZONTAL" ? lastWord.x + i : lastWord.x,
            y: lastWord.direction === "VERTICAL" ? lastWord.y + i : lastWord.y,
        }));
    });

    function handleAnnouncementTap() {
        const at = game.announcement?.at;
        if (at) {
            boardComponent?.centerOn(at.x, at.y);
        }
        game.announcement = null;
    }

    let canVote = $derived(
        game.challenge !== null
        && !game.myVote
        && game.playerId !== null
        && game.playerId !== game.challenge?.moverId
        && game.playerId !== game.challenge?.challengerId
    );
</script>

<svelte:window onresize={applyWindowSize} />
<svelte:head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no" />
</svelte:head>

<style>
    .panel {
        display: flex;
        flex-direction: column;
        align-items: center;
        position: fixed;
        width: 100%;
        box-sizing: border-box;
        background-color: rgba(255,255,255,0.5);
        backdrop-filter: blur(10px);
        box-shadow: 0 0 10px rgba(0,0,0,0.1);
        padding: 1rem;
        gap: 0.5rem;
    }

    input {
        display: block;
        box-sizing: border-box;
        width: 100%;
        outline: none;
        border: 1px solid rgba(0, 0, 0, 0.25);
        border-radius: 0.5rem;
        background-color: rgba(255, 255, 255, 0.75);
        padding: 0.5rem 0.75rem;
        font: inherit;
        text-align: center;
        margin-bottom: 0.5rem;
    }

    .input {
        margin-top: 0.5rem;
        margin-bottom: 0;
    }

    header {
        top: 0;
        padding: 0.5rem 1rem;
    }

    footer {
        bottom: 0;
        justify-content: center;
    }

    .scores {
        display: flex;
        flex-wrap: wrap;
        justify-content: center;
        gap: 0.25rem 1rem;
        margin: 0;
        padding: 0;
        list-style: none;
        font-size: 0.9rem;
    }

    .scores .current {
        font-weight: bold;
    }

    .status {
        margin: 0;
        font-size: 0.85rem;
        color: rgba(0,0,0,0.6);
    }

    .announcement {
        font: inherit;
        font-size: 0.85rem;
        padding: 0.35rem 0.9rem;
        border-radius: 999px;
        border: 1px solid rgba(37,99,235,0.4);
        background-color: rgba(37,99,235,0.12);
        color: #1e40af;
        cursor: pointer;
    }

    .recenter {
        position: fixed;
        right: 1rem;
        bottom: 40%;
        font: inherit;
        font-size: 0.85rem;
        padding: 0.5rem 1rem;
        border-radius: 999px;
        border: 1px solid rgba(0,0,0,0.2);
        background-color: rgba(255,255,255,0.9);
        box-shadow: 0 2px 8px rgba(0,0,0,0.15);
        cursor: pointer;
    }

    .actions {
        display: flex;
        flex-wrap: wrap;
        justify-content: center;
        gap: 0.5rem;
    }

    .actions button {
        font: inherit;
        font-size: 0.9rem;
        padding: 0.4rem 0.9rem;
        border-radius: 0.5rem;
        border: 1px solid rgba(0,0,0,0.25);
        background-color: rgba(255,255,255,0.85);
        cursor: pointer;
    }

    .actions button.primary {
        background-color: #2563eb;
        border-color: #2563eb;
        color: white;
    }

    .actions button:disabled {
        opacity: 0.5;
        cursor: default;
    }

    .error {
        margin: 0;
        color: #b91c1c;
        font-size: 0.9rem;
        text-align: center;
    }

    .hint {
        margin: 0;
        font-size: 0.85rem;
        color: rgba(0,0,0,0.55);
        text-align: center;
    }

    .challenge {
        text-align: center;
        font-size: 0.9rem;
    }

    .challenge p {
        margin: 0 0 0.5rem;
    }

    .gameover {
        position: fixed;
        inset: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        background-color: rgba(255,255,255,0.75);
        backdrop-filter: blur(4px);
    }

    .gameover > div {
        background: white;
        border-radius: 1rem;
        box-shadow: 0 10px 30px rgba(0,0,0,0.15);
        padding: 2rem;
        text-align: center;
        max-width: 20rem;
    }

    .gameover h2 {
        margin-top: 0;
    }

    .gameover ol {
        list-style: none;
        margin: 0;
        padding: 0;
    }
</style>

{#if !game.loaded}
    <p>Loading game...</p>
{:else}
    <Board
            bind:this={boardComponent}
            cells={[...game.board.cells]}
            requestCells={(x1, y1, x2, y2) => game.loadBoard(x1, y1, x2, y2)}
            onCellTap={handleCellTap}
            onViewChange={(view) => boardView = view}
            ghostCells={ghostCells}
            highlightCell={selectedCell}
            highlightCells={lastWordCells}
            width={boardWidth}
            height={boardHeight}
            offsetX={offsetX}
            offsetY={offsetY}
            letterPoints={game.letterPoints}
            style="position: fixed; top: 0; left: 0; width: 100%; height: 100%;"
            cellSize={50}
    />

    {#if showRecenter}
        <button class="recenter" onclick={() => boardComponent?.centerOn(0, 0)}>
            Back to center
        </button>
    {/if}

    {#if game.started}
        <header class="panel">
            <ul class="scores">
                {#each game.players as player (player.id)}
                    <li class:current={player.id === game.currentPlayerId}>
                        {player.name}{player.id === game.playerId ? " (you)" : ""}: {player.score}
                    </li>
                {/each}
            </ul>
            {#if !game.finished}
                <p class="status">
                    {game.myTurn ? "Your turn" : `${game.playerName(game.currentPlayerId)}'s turn`}
                    · {game.lettersRemaining} letters left
                </p>
            {/if}
            {#if game.announcement}
                <button class="announcement" onclick={handleAnnouncementTap}>
                    {game.announcement.text}{game.announcement.at ? " · tap to view" : ""}
                </button>
            {/if}
        </header>
    {/if}

    <footer class="panel">
        {#if game.error}
            <p class="error">{game.error}</p>
        {/if}

        {#if !game.playerId}
            <div>
                <label>
                    Name
                    <input type="text" bind:value={name}>
                </label>

                <button class="button" onclick={async () => await game.join(name)}>Join game</button>
            </div>
        {:else if !game.started}
            <div class="actions">
                <button class="primary" onclick={(e) => {
                    e.preventDefault();

                    game.start();
                }}>Start game</button>
            </div>
            <p class="hint">Waiting for players... share this page's link to invite them.</p>
        {:else if game.challenge}
            <div class="challenge">
                <p>
                    <strong>{game.playerName(game.challenge.challengerId)}</strong> challenged
                    <strong>{game.playerName(game.challenge.moverId)}</strong>'s word!
                    ({game.challenge.votesInvalid + game.challenge.votesValid} of {game.challenge.eligibleVoters} votes in)
                </p>
                {#if canVote}
                    <div class="actions">
                        <button onclick={() => game.castVote("VALID")}>Real word</button>
                        <button onclick={() => game.castVote("INVALID")}>Not a word</button>
                    </div>
                {:else}
                    <p class="hint">Waiting for votes...</p>
                {/if}
            </div>
        {:else if selectedPlacement}
            <div class="actions">
                {#if game.placements.length > 1}
                    <button onclick={() => game.selectedPlacement = (game.selectedPlacement + game.placements.length - 1) % game.placements.length}>&larr;</button>
                {/if}
                <button class="primary" onclick={playWord}>
                    Play {selectedPlacement.word} for {selectedPlacement.points} pts
                    {#if game.placements.length > 1}
                        ({game.selectedPlacement + 1}/{game.placements.length})
                    {/if}
                </button>
                {#if game.placements.length > 1}
                    <button onclick={() => game.selectedPlacement = (game.selectedPlacement + 1) % game.placements.length}>&rarr;</button>
                {/if}
                <button onclick={cancelPlacement}>Cancel</button>
            </div>
        {:else}
            <div style="width: 100%; max-width: 24rem;">
                <Rack
                        letters={game.sortedRack}
                        letterPoints={game.letterPoints}
                        input={game.input}
                        onTapLetter={(letter, used) => game.tapLetter(letter, used)}
                        onReorder={(letters) => game.setRackOrder(letters)}
                />
                {#if !game.finished}
                    <input type="text" bind:value={game.input} placeholder="WORD" class="input" />
                    <p class="hint">
                        {#if !game.myTurn}
                            Plan your next word while you wait.
                        {:else if game.board.cells.some(cell => cell.letter)}
                            Type a word, then tap the square where it starts.
                        {:else}
                            Type a word, then tap the center star to place it.
                        {/if}
                    </p>
                {/if}
            </div>
            <div class="actions">
                {#if game.myTurn}
                    <button onclick={() => game.pass()}>Skip my turn</button>
                    <button
                            disabled={game.input.length === 0}
                            onclick={() => game.exchange([...game.input])}
                    >{game.input.length > 0 ? `Swap ${game.input.length} letters` : "Swap letters"}</button>
                {/if}
                {#if canChallenge && game.challengeableMoverId}
                    <button onclick={() => game.challengeWord()}>Challenge {game.playerName(game.challengeableMoverId)}'s word</button>
                {/if}
            </div>
        {/if}
    </footer>

    {#if game.finished}
        <div class="gameover">
            <div>
                <h2>
                    {#if game.winnerIds.length === 1}
                        {game.playerName(game.winnerIds[0])} wins!
                    {:else if game.winnerIds.length > 1}
                        It's a tie!
                    {:else}
                        Game over
                    {/if}
                </h2>
                <ol>
                    {#each [...game.players].sort((a, b) => b.score - a.score) as player (player.id)}
                        <li>{player.name}: {player.score}</li>
                    {/each}
                </ol>
            </div>
        </div>
    {/if}
{/if}
