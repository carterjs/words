<script lang="ts">
    import { untrack } from "svelte";
    import { flip } from "svelte/animate";
    import Tile from "$lib/Tile.svelte";

    type Props = {
        letters: string[];
        letterPoints?: Record<string, number>;
        input?: string;
        onTapLetter?: (letter: string, used: boolean) => void;
        onReorder?: (letters: string[]) => void;
    }

    const { letters, letterPoints={}, input="", onTapLetter, onReorder }: Props = $props();

    // tiles carry stable ids so reorders can animate; reconcile against the
    // letters prop, reusing ids for letters still on the rack
    let tiles = $state<{ id: number; letter: string }[]>([]);
    let nextId = 0;

    $effect(() => {
        const incoming = letters;
        untrack(() => {
            const pool = [...tiles];
            tiles = incoming.map((letter) => {
                const index = pool.findIndex(tile => tile.letter === letter);
                if (index !== -1) {
                    return pool.splice(index, 1)[0];
                }
                return { id: nextId++, letter };
            });
        });
    });

    // the tiles the player explicitly tapped into the word, so duplicates
    // highlight the copy that was actually touched
    let tappedIds = $state<number[]>([]);

    $effect(() => {
        if (input === "") {
            tappedIds = [];
        }
    });

    // tiles spent on the current word, highlighted in place: tapped tiles
    // first, then first-match for typed letters
    let usedIds = $derived.by(() => {
        const ids = new Set<number>();
        const needed = new Map<string, number>();
        for (const letter of input) {
            needed.set(letter, (needed.get(letter) ?? 0) + 1);
        }

        for (const id of tappedIds) {
            const tile = tiles.find(candidate => candidate.id === id);
            if (tile && (needed.get(tile.letter) ?? 0) > 0) {
                ids.add(tile.id);
                needed.set(tile.letter, (needed.get(tile.letter) ?? 0) - 1);
            }
        }

        for (const tile of tiles) {
            if (!ids.has(tile.id) && (needed.get(tile.letter) ?? 0) > 0) {
                ids.add(tile.id);
                needed.set(tile.letter, (needed.get(tile.letter) ?? 0) - 1);
            }
        }

        return ids;
    });

    let listElement: HTMLUListElement;

    // index within tiles of the one being dragged, if any
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

        const target = Math.max(0, Math.min(tiles.length - 1, nearestDisplayIndex(e)));
        if (target !== dragIndex) {
            const next = [...tiles];
            const [moved] = next.splice(dragIndex, 1);
            next.splice(target, 0, moved);
            tiles = next;
            dragIndex = target;
        }
    }

    function handlePointerUp() {
        if (dragIndex === null) return;
        const tile = tiles[dragIndex];

        if (dragMoved) {
            onReorder?.(tiles.map(t => t.letter));
        } else if (tile) {
            const used = usedIds.has(tile.id);
            if (used) {
                tappedIds = tappedIds.filter(id => id !== tile.id);
            } else {
                tappedIds = [...tappedIds, tile.id];
            }
            onTapLetter?.(tile.letter, used);
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
    {#each tiles as tile, index (tile.id)}
        <li
                animate:flip={{ duration: 150 }}
                class:dragging={dragIndex === index && dragMoved}
                onpointerdown={(e) => handlePointerDown(e, index)}
                onpointermove={handlePointerMove}
                onpointerup={handlePointerUp}
                onpointercancel={handlePointerCancel}
        >
            <Tile cellSize={50} letter={tile.letter} x={0} y={0} points={letterPoints[tile.letter]} selected={usedIds.has(tile.id)} />
        </li>
    {/each}
</ul>

{#if dragIndex !== null && dragMoved && tiles[dragIndex]}
    <div class="floating" use:portal style="left: {dragPosition.x - 25}px; top: {dragPosition.y - dragLift}px;">
        <Tile cellSize={50} letter={tiles[dragIndex].letter} x={0} y={0} points={letterPoints[tiles[dragIndex].letter]} selected />
    </div>
{/if}
