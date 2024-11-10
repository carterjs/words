<script lang="ts">
    import {GameController} from '$lib/game.svelte.ts';
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
        game.input = game.input.toUpperCase();
    })
</script>

<svelte:window onresize={applyWindowSize} />
<svelte:head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no" />
</svelte:head>

<style>
    main {
        padding: 0;
        max-width: 960px;
        margin: 0 auto;
        display: flex;
        flex-direction: column;
        height: 100%;
    }

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
    }

    footer {
        bottom: 0;
        justify-content: center;
    }
</style>

{#if !game.loaded}
    <p>Loading game...</p>
{:else}
    <Board
            cells={[...game.board.cells]}
            requestCells={(x1, y1, x2, y2) => game.loadBoard(x1, y1, x2, y2)}
            width={boardWidth}
            height={boardHeight}
            offsetX={offsetX}
            offsetY={offsetY}
            letterPoints={game.letterPoints}
            style="position: fixed; top: 0; left: 0; width: 100%; height: 100%;"
            cellSize={50}
    />
    <div class="panel">
        {#if !game.playerId}
            <div>
                <label>
                    Name
                    <input type="text" bind:value={name}>
                </label>

                <button class="button" onclick={async () => await game.join(name)}>Join game</button>
            </div>
        {:else}
            {#if game.started}
                <div>
                    <Rack letters={game.sortedRack} letterPoints={game.letterPoints} input={game.input} />
                    <input type="text" bind:value={game.input} placeholder="WORD" class="input" />
                </div>
            {:else}
                <div>
                    <button class="button" onclick={(e) => {
                        e.preventDefault();

                        game.start();
                    }}>Start game</button>
                </div>
            {/if}
        {/if}
    </div>
{/if}
