<script lang="ts">
    import type {Cell} from "$lib/game.svelte";
    import Tile from "$lib/Tile.svelte";
    import Modifier from "$lib/Modifier.svelte";

    type Props = {
        letterPoints?: Record<string, number>;
        cells: Cell[];
        requestCells?: (x1: number, y1: number, x2: number, y2: number) => void;
        onCellTap?: (x: number, y: number) => void;
        ghostCells?: Cell[];
        highlightCell?: { x: number; y: number } | null;
        highlightCells?: { x: number; y: number }[];
        width: number;
        height: number;
        scale?: number;
        offsetX?: number;
        offsetY?: number;
        cellSize?: number;
        style?: string;
        disabled?: boolean;
    }

    let {
        letterPoints = {},
        cells = [],
        requestCells = (x1, y1, x2, y2) => console.log(x1, y1, x2, y2),
        onCellTap,
        ghostCells = [],
        highlightCell = null,
        highlightCells = [],
        width,
        height,
        scale: initialScale = 1,
        offsetX = 0,
        offsetY = 0,
        cellSize = 60,
        style = "",
        disabled
    }: Props = $props()

    let scale = $state(initialScale)

    const minScale = 0.35;
    const maxScale = 2.5;

    let anchor: null | { x: number; y: number } = null;

    let displacementX = $state(0);
    let displacementY = $state(0);

    let minX = $derived(Math.min(...cells.map(cell => cell.x)));
    let minY = $derived(Math.min(...cells.map(cell => cell.y)));
    let maxX = $derived(Math.max(...cells.map(cell => cell.x)));
    let maxY = $derived(Math.max(...cells.map(cell => cell.y)));

    let svgElement: SVGSVGElement;

    // offsetX/offsetY are relative to the event target, which can be a tile
    // or the center star rather than the svg - measure against the svg itself
    function pointerPosition(e: { clientX: number; clientY: number }) {
        const rect = svgElement.getBoundingClientRect();
        return { x: e.clientX - rect.left, y: e.clientY - rect.top };
    }

    let pointers = new Map<number, { x: number; y: number }>();
    let pinched = false;

    function pointerDistance() {
        const [first, second] = [...pointers.values()];
        return Math.hypot(first.x - second.x, first.y - second.y);
    }

    function pointerMidpoint() {
        const [first, second] = [...pointers.values()];
        return { x: (first.x + second.x) / 2, y: (first.y + second.y) / 2 };
    }

    function clampScale(value: number) {
        return Math.min(maxScale, Math.max(minScale, value));
    }

    // centerOn pans the view so the given cell sits at the center of the board
    export function centerOn(cellX: number, cellY: number) {
        offsetX = (cellX + 0.5) * cellSize * scale - width / 2;
        offsetY = (cellY + 0.5) * cellSize * scale - height / 2;
    }

    // zoomTo changes the scale while keeping the board point under the given
    // screen position fixed (and optionally following it to a new position)
    function zoomTo(nextScale: number, from: { x: number; y: number }, to = from) {
        offsetX = (offsetX + from.x) * (nextScale / scale) - to.x;
        offsetY = (offsetY + from.y) * (nextScale / scale) - to.y;
        scale = nextScale;
    }

    function handlePointerDown(e: PointerEvent) {
        if (disabled) return;
        const position = pointerPosition(e);
        pointers.set(e.pointerId, position);
        try {
            svgElement.setPointerCapture?.(e.pointerId);
        } catch {
            // synthetic events have no active pointer to capture
        }

        if (pointers.size === 2) {
            // second finger down: commit any in-progress pan, start pinching
            offsetX += displacementX;
            offsetY += displacementY;
            displacementX = 0;
            displacementY = 0;
            anchor = null;
            pinched = true;
        } else if (pointers.size === 1) {
            anchor = position;
            pinched = false;
        } else {
            anchor = null;
        }
    }

    function handlePointerMove(e: PointerEvent) {
        if (!pointers.has(e.pointerId)) return;

        if (pointers.size === 2) {
            const previousMidpoint = pointerMidpoint();
            const previousDistance = pointerDistance();
            pointers.set(e.pointerId, pointerPosition(e));

            if (previousDistance > 0) {
                zoomTo(clampScale(scale * (pointerDistance() / previousDistance)), previousMidpoint, pointerMidpoint());
            }
            return;
        }

        const position = pointerPosition(e);
        pointers.set(e.pointerId, position);

        if (anchor) {
            displacementX = anchor.x - position.x;
            displacementY = anchor.y - position.y;
        }
    }

    function handlePointerUp(e?: PointerEvent) {
        if (e) {
            pointers.delete(e.pointerId);
        } else {
            pointers.clear();
        }

        // a pointer that barely moved is a tap on a cell, not a pan or pinch
        const wasTap = e && anchor && !pinched
            && Math.abs(displacementX) < 8
            && Math.abs(displacementY) < 8;

        anchor = null;
        offsetX += displacementX;
        offsetY += displacementY;
        displacementX = 0;
        displacementY = 0;

        if (wasTap && onCellTap) {
            const position = pointerPosition(e);
            const cellX = Math.floor((offsetX + position.x) / scale / cellSize);
            const cellY = Math.floor((offsetY + position.y) / scale / cellSize);
            onCellTap(cellX, cellY);
        }
    }

    // Svelte registers template wheel handlers as passive, so attach directly
    // to be able to preventDefault the page zoom/scroll
    $effect(() => {
        const handleWheel = (e: WheelEvent) => {
            if (disabled) return;
            e.preventDefault();
            zoomTo(clampScale(scale * Math.exp(-e.deltaY * 0.002)), pointerPosition(e));
        };

        svgElement.addEventListener("wheel", handleWheel, { passive: false });
        return () => svgElement.removeEventListener("wheel", handleWheel);
    })

    $effect(() => {
        let newMinX = Math.min(minX, Math.floor(offsetX / scale / cellSize) - 5);
        let newMinY = Math.min(minY, Math.floor(offsetY / scale / cellSize) - 5);
        let newMaxX = Math.max(maxX, Math.ceil((offsetX + width) / scale / cellSize) + 5);
        let newMaxY = Math.max(maxY, Math.ceil((offsetY + height) / scale / cellSize) + 5);

        if (newMinX != minX) {
            requestCells(newMinX, newMinY, minX, newMaxY);
        }

        if (newMaxX != maxX) {
            requestCells(maxX, newMinY, newMaxX, newMaxY);
        }

        if (newMinY != minY) {
            requestCells(minX, newMinY, maxX, minY);
        }

        if (newMaxY != maxY) {
            requestCells(minX, maxY, maxX, newMaxY);
        }
    })

    let visibleCells = $derived.by(() => {
        const left = (offsetX + displacementX) / scale;
        const top = (offsetY + displacementY) / scale;
        const right = left + width / scale;
        const bottom = top + height / scale;

        return cells.filter(cell => {
            if (cell.x * cellSize + cellSize < left - 5 * cellSize) return false;
            if (cell.y * cellSize + cellSize < top - 5 * cellSize) return false;
            if (cell.x * cellSize > right + 5 * cellSize) return false;
            if (cell.y * cellSize > bottom + 5 * cellSize) return false;

            return true;
        });
    })

    const precision = 2;

    // this is overkill but fun
    const points = 5;
    const innerRadius = cellSize/7;
    const outerRadius = cellSize/2.5;
    const slice = Math.PI * 2 / points;
    let starPoints = $derived(Array(points).fill(0).map((_, i) => {
        return [
            // outer point
            `${Math.sin(i*slice - Math.PI) * outerRadius + cellSize/2} ${Math.cos(i*slice - Math.PI) * outerRadius + cellSize/2}`,
            // inner point
            `${Math.sin((i+0.5)*slice - Math.PI) * innerRadius + cellSize/2} ${Math.cos((i+0.5)*slice - Math.PI) * innerRadius + cellSize/2}`
        ]
    }).flat().join(","))

    console.log("sp", starPoints)
</script>

<style>
    svg :global(*) {
        pointer-events: none;
    }

    svg {
        user-select: none;
        touch-action: none;
        cursor: pointer;
        background-position: var(--offset-x) var(--offset-y);
        background-size: var(--cell-size) var(--cell-size);
        background-image:
                linear-gradient(to right, rgba(0,0,0,0.1) 1px, transparent 1px),
                linear-gradient(to bottom, rgba(0,0,0,0.1) 1px, transparent 1px);
        display: block;
        font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
    }

    @keyframes fade-in {
        from {
            opacity: 0;
        }
        to {
            opacity: 1;
        }
    }
</style>

<svg
        bind:this={svgElement}
        style="--cell-size: {cellSize*scale}px; --offset-x: {Math.round(((-offsetX-displacementX-0.5)%(cellSize*scale))*precision)/precision}px; --offset-y: {Math.round(((-offsetY-displacementY-0.5)%(cellSize*scale))*precision)/precision}px;{style}"
        viewBox={`${Math.round((offsetX+displacementX) / scale * precision) / precision} ${Math.round((offsetY+displacementY) / scale * precision) / precision} ${Math.round(width / scale * precision) / precision} ${Math.round(height / scale*precision)/precision}`}
        width={width}
        height={height}
        onpointerdown={handlePointerDown}
        onpointermove={handlePointerMove}
        onpointerup={handlePointerUp}
        onpointercancel={() => handlePointerUp()}
        onpointerleave={() => handlePointerUp()}
>
    {#if !cells.some(cell => cell.x === 0 && cell.y === 0 && cell.letter)}
        <rect
                x={0}
                y={0}
                width={cellSize}
                height={cellSize}
                fill="rgba(255,255,255,0.75)"
                stroke="rgba(0,0,0,0.125)"
        />
        <polygon
                points={starPoints}
                fill="rgba(0,255,0,0.25)"
                stroke="rgba(0,0,0,0.25)"
        />
    {/if}
    {#each visibleCells as cell (cell)}
        {#if cell.modifier}
            <Modifier
                cellSize={cellSize}
                x={cell.x*cellSize}
                y={cell.y*cellSize}
                modifier={cell.modifier}
            />
        {/if}
        {#if cell.letter}
            <Tile
                cellSize={cellSize}
                x={cell.x*cellSize}
                y={cell.y*cellSize}
                letter={cell.letter}
                points={letterPoints[cell.letter]}
            />
        {/if}
    {/each}
    {#each highlightCells as cell (cell)}
        <rect
                class="highlight"
                x={cell.x*cellSize}
                y={cell.y*cellSize}
                width={cellSize}
                height={cellSize}
                fill="rgba(37,99,235,0.18)"
                stroke="rgba(37,99,235,0.4)"
                stroke-width="1.5"
        />
    {/each}
    {#if highlightCell}
        <rect
                x={highlightCell.x*cellSize}
                y={highlightCell.y*cellSize}
                width={cellSize}
                height={cellSize}
                fill="rgba(37,99,235,0.15)"
                stroke="rgba(37,99,235,0.6)"
                stroke-width="2"
        />
    {/if}
    {#each ghostCells as cell (cell)}
        {#if cell.letter}
            <g opacity="0.65">
                <Tile
                    cellSize={cellSize}
                    x={cell.x*cellSize}
                    y={cell.y*cellSize}
                    letter={cell.letter}
                    points={letterPoints[cell.letter]}
                    selected
                />
            </g>
        {/if}
    {/each}
</svg>