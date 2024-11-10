<script lang="ts">
    import type {Cell} from "$lib/game.svelte";
    import Tile from "$lib/Tile.svelte";
    import Modifier from "$lib/Modifier.svelte";

    type Props = {
        letterPoints?: Record<string, number>;
        cells: Cell[];
        requestCells?: (x1: number, y1: number, x2: number, y2: number) => void;
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

    let anchor: null | { x: number; y: number } = null;

    let displacementX = $state(0);
    let displacementY = $state(0);

    let minX = $derived(Math.min(...cells.map(cell => cell.x)));
    let minY = $derived(Math.min(...cells.map(cell => cell.y)));
    let maxX = $derived(Math.max(...cells.map(cell => cell.x)));
    let maxY = $derived(Math.max(...cells.map(cell => cell.y)));

    function handlePointerDown(e: PointerEvent) {
        if (disabled) return;
        anchor = { x: e.offsetX, y: e.offsetY };
    }

    function handlePointerMove(e: PointerEvent) {
        if (anchor) {
            displacementX = anchor.x - e.offsetX;
            displacementY =  anchor.y - e.offsetY;
        }
    }

    function handlePointerUp() {
        anchor = null;
        offsetX += displacementX;
        offsetY += displacementY;
        displacementX = 0;
        displacementY = 0;
    }

    $effect(() => {
        let newMinX = Math.min(minX, Math.floor(offsetX / cellSize) - 5);
        let newMinY = Math.min(minY, Math.floor(offsetY / cellSize) - 5);
        let newMaxX = Math.max(maxX, Math.ceil((offsetX + width) / cellSize) + 5);
        let newMaxY = Math.max(maxY, Math.ceil((offsetY + height) / cellSize) + 5);

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

    let visibleCells = $derived(cells.filter(cell => {
        if (cell.x * cellSize + cellSize < offsetX + displacementX - 5 * cellSize) return false;
        if (cell.y * cellSize + cellSize < offsetY + displacementY - 5 * cellSize) return false;
        if (cell.x * cellSize > offsetX + displacementX + width + 5 * cellSize) return false;
        if (cell.y * cellSize > offsetY + displacementY + height + 5 * cellSize) return false;

        return true;
    }))

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
    svg {
        user-select: none;
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
        style="--cell-size: {cellSize*scale}px; --offset-x: {Math.round((-offsetX-displacementX-0.5)%cellSize*precision)/precision}px; --offset-y: {Math.round((-offsetY-displacementY-0.5)%cellSize*precision)/precision}px;{style}"
        viewBox={`${Math.round((offsetX+displacementX) / scale * precision) / precision} ${Math.round((offsetY+displacementY) / scale * precision) / precision} ${Math.round(width / scale * precision) / precision} ${Math.round(height / scale*precision)/precision}`}
        width={width}
        height={height}
        onpointerdown={handlePointerDown}
        onpointermove={handlePointerMove}
        onpointerup={handlePointerUp}
        onpointercancel={handlePointerUp}
        onpointerleave={handlePointerUp}
>
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
</svg>