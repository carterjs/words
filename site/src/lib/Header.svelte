<script>
    import Board from "$lib/Board.svelte";
</script>
<style>
    header {
        display: flex;
        justify-content: center;
        padding-bottom: 0;
    }
</style>
<header>
    <a href="/" title="Words">
        <Board
                offsetX={-15}
                offsetY={-15}
                cellSize={60}
                width={60*5+30}
                height={60+30}
                cells={[
                    { x: 0, y: 0, letter: "W" },
                    { x: 1, y: 0, letter: "O" },
                    { x: 2, y: 0, letter: "R" },
                    { x: 3, y: 0, letter: "D" },
                    { x: 4, y: 0, letter: "S" },
                ]}
                letterPoints={{
                    "W": 4,
                    "O": 1,
                    "R": 1,
                    "D": 2,
                    "S": 1,
                }}
                style="border: 3px solid transparent; touch-action: none;"
                disabled
        />
    </a>
</header>