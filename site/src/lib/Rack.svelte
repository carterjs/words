<script lang="ts">
    import Tile from "$lib/Tile.svelte";

    type Props = {
        letters: string[];
        letterPoints?: Record<string, number>;
        input?: string;
        onTapLetter?: (letter: string, used: boolean) => void;
        onReorder?: (letters: string[]) => void;
    }

    const { letters, letterPoints={}, input="", onTapLetter, onReorder }: Props = $props();

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

    let listElement: HTMLUListElement;

    // index within unusedLetters of the tile being dragged, if any
    let dragIndex = $state<number | null>(null);
    let dragMoved = $state(false);
    let dragStart = { x: 0, y: 0 };

    // where the dragged tile floats; lifted above the finger on touch so it
    // isn't hidden under it
    let dragPosition = $state({ x: 0, y: 0 });
    let dragLift = $state(0);

    // the tile slot closest to the pointer, robust to flex wrapping
    function nearestDisplayIndex(e: PointerEvent) {
        let best = 0;
        let bestDistance = Infinity;

        [...listElement.children].forEach((child, index) => {
            const rect = child.getBoundingClientRect();
            const distance = Math.hypot(
                e.clientX - (rect.left + rect.width / 2),
                e.clientY - (rect.top + rect.height / 2),
            );
            if (distance < bestDistance) {
                bestDistance = distance;
                best = index;
            }
        });

        return best;
    }

    function handlePointerDown(e: PointerEvent, index: number) {
        dragIndex = index;
        dragMoved = false;
        dragStart = { x: e.clientX, y: e.clientY };
        dragPosition = { x: e.clientX, y: e.clientY };
        dragLift = e.pointerType === "touch" ? 65 : 25;
        try {
            (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
        } catch {
            // synthetic events have no active pointer to capture
        }
    }

    function handlePointerMove(e: PointerEvent) {
        if (dragIndex === null) return;

        dragPosition = { x: e.clientX, y: e.clientY };

        if (!dragMoved && Math.hypot(e.clientX - dragStart.x, e.clientY - dragStart.y) > 8) {
            dragMoved = true;
        }
        if (!dragMoved) return;

        const target = Math.max(0, Math.min(
            unusedLetters.length - 1,
            nearestDisplayIndex(e) - usedLetters.length,
        ));

        if (target !== dragIndex) {
            const next = [...unusedLetters];
            const [moved] = next.splice(dragIndex, 1);
            next.splice(target, 0, moved);
            unusedLetters = next;
            dragIndex = target;
        }
    }

    function handlePointerUp() {
        if (dragIndex === null) return;

        if (dragMoved) {
            onReorder?.([...usedLetters, ...unusedLetters]);
        } else {
            onTapLetter?.(unusedLetters[dragIndex], false);
        }

        dragIndex = null;
        dragMoved = false;
    }

    function handlePointerCancel() {
        dragIndex = null;
        dragMoved = false;
    }

    // ancestors with backdrop-filter (the footer panel) become the containing
    // block for position:fixed, throwing off viewport coordinates - render
    // the floating tile from the body instead
    function portal(node: HTMLElement) {
        document.body.appendChild(node);
        return {
            destroy() {
                node.remove();
            },
        };
    }
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
        touch-action: none;
        user-select: none;
        -webkit-user-select: none;
        cursor: pointer;
    }

    li.dragging {
        opacity: 0.25;
    }

    .floating {
        position: fixed;
        z-index: 100;
        pointer-events: none;
        transform: scale(1.2);
        filter: drop-shadow(0 6px 10px rgba(0,0,0,0.3));
    }
</style>

<ul bind:this={listElement}>
    {#each usedLetters as letter, index (index)}
        <li onpointerup={() => onTapLetter?.(letter, true)}>
            <Tile cellSize={50} letter={letter} x={0} y={0} points={letterPoints[letter]} selected />
        </li>
    {/each}
    {#each unusedLetters as letter, index (index)}
        <li
                class:dragging={dragIndex === index}
                onpointerdown={(e) => handlePointerDown(e, index)}
                onpointermove={handlePointerMove}
                onpointerup={handlePointerUp}
                onpointercancel={handlePointerCancel}
        >
            <Tile cellSize={50} letter={letter} x={0} y={0} points={letterPoints[letter]} />
        </li>
    {/each}
</ul>

{#if dragIndex !== null && dragMoved}
    <div class="floating" use:portal style="left: {dragPosition.x - 25}px; top: {dragPosition.y - dragLift}px;">
        <Tile cellSize={50} letter={unusedLetters[dragIndex]} x={0} y={0} points={letterPoints[unusedLetters[dragIndex]]} selected />
    </div>
{/if}
