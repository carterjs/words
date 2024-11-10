<script lang="ts">
    import Tile from "$lib/Tile.svelte";

    type Props = {
        letters: string[];
        setLetters?: (letters: string[]) => void;
        letterPoints?: Record<string, number>;
        input?: string;
    }

    const { letters, letterPoints={}, input="" }: Props = $props();

    let unusedLetters = $state<string[]>([]);
    let usedLetters = $state<string[]>([]);

    $effect(() => {
        let nextUnusedLetters = [...letters];
        let nextUsedLetters = [];

        for (let letter of input) {
            let index = nextUnusedLetters.indexOf(letter);
            if (index !== -1) {
                nextUnusedLetters.splice(index, 1);
                nextUsedLetters.push(letter);
            }
        }

        unusedLetters = nextUnusedLetters;
        usedLetters = nextUsedLetters;
    })
</script>

<style>
    ul {
        display: flex;
        justify-content: center;
        flex-wrap: wrap;
        list-style: none;
        padding: 0;
        margin: 0;
    }

    li {
        display: block;
        padding: 0;
        margin: 0;
    }

</style>

<ul>
    {#each usedLetters as letter}
        <li>
            <Tile cellSize={50} letter={letter} x={0} y={0} points={letterPoints[letter]} selected />
        </li>
    {/each}
    {#each unusedLetters as letter}
        <li>
            <Tile cellSize={50} letter={letter} x={0} y={0} points={letterPoints[letter]} />
        </li>
    {/each}
</ul>