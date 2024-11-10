<script lang="ts">
    import {GameController} from '$lib/game.svelte.ts';
    import { page } from '$app/stores';
    import {onMount} from "svelte";
    import {PUBLIC_API_URL} from "$env/static/public";
    import Board from "$lib/Board.svelte";

    console.log(PUBLIC_API_URL);

    let gameId = $state<string>("");
    let game = new GameController();

    let boardWidth = $state(0);
    let boardHeight = $state(0);
    let offsetX = $state(0);
    let offsetY = $state(0);

    onMount(async () => {
        gameId = $page.url.searchParams.get("game") || "";
        if (!gameId) {
            return
        }

        game.id = gameId;

        await game.loadInitialData();
        game.streamUpdates()

        applyWindowSize();
        offsetY = -boardHeight / 2 + 30;
        offsetX = -boardWidth / 2 + 30;
    });

    function applyWindowSize() {
        boardWidth = window.innerWidth;
        boardHeight = window.innerHeight;
    }

    let name = $state("");
</script>

<svelte:window onresize={applyWindowSize} />

<style>
    main {
        padding: 0;
        margin: 0;
        max-width: 960px;
        margin: 0 auto;
        display: flex;
        flex-direction: column;
        height: 100%;
    }
</style>

<main>
    {#if !game.loaded}
        <p>Loading game...</p>
    {:else}
        {#if !game.playerId}
            <label>
                Name:
                <input type="text" bind:value={name}>
            </label>

            <button onclick={async () => await game.join(name)}>Join game as {name}</button>
        {:else}
            <h1>Hello, {game.getPlayerById(game.playerId)?.name}</h1>
            {#if !game.started}
                <button onclick={async () => await game.start()}>Start game</button>
            {/if}
            <p>
                {#each game.rack as letter}
                    <span>{letter}</span>
                {/each}
            </p>
            <Board
                cells={[...game.board.cells, {x: 0, y: 0, letter: "A"}]}
                requestCells={(x1, y1, x2, y2) => game.loadBoard(x1, y1, x2, y2)}
                width={boardWidth}
                height={boardHeight}
                offsetX={offsetX}
                offsetY={offsetY}
                style="position: fixed; top: 0; left: 0; width: 100%; height: 100%;"
            />
            <button>test</button>
        {/if}
    {/if}
</main>