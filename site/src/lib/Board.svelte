<script lang="ts">
    import type {Cell} from "$lib/game.svelte";

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
        let newMinX = Math.min(minX, Math.floor(offsetX / cellSize));
        let newMinY = Math.min(minY, Math.floor(offsetY / cellSize));
        let newMaxX = Math.max(maxX, Math.ceil((offsetX + width) / cellSize));
        let newMaxY = Math.max(maxY, Math.ceil((offsetY + height) / cellSize));

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
</script>

<style>
    svg {
        cursor: pointer;
        background-position: var(--offset-x) var(--offset-y);
        background-size: var(--cell-size) var(--cell-size);
        background-image:
                linear-gradient(to right, rgba(0,0,0,0.1) 1px, transparent 1px),
                linear-gradient(to bottom, rgba(0,0,0,0.1) 1px, transparent 1px);
        display: block;
        font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
    }

    .tile {
        filter: drop-shadow( 2px 3px 5px rgba(0, 0, 0, .25));
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
    {#each visibleCells as { x, y, letter, modifier }}
        {#if modifier}
            <g>
                <rect
                        x={x*cellSize}
                        y={y*cellSize}
                        width={cellSize}
                        height={cellSize}
                        fill={{
                                "DW": "rgba(255,170,255,0.75)",
                                "DL": "rgba(170,170,255,0.75)",
                                "TW": "rgba(255,170,170,0.75)",
                                "TL": "rgba(255,170,170,0.75)",
                            }[modifier]}
                        stroke="none"
                />
                <text
                        x={x*cellSize + cellSize/2}
                        y={y*cellSize + cellSize/2}
                        text-anchor="middle"
                        dominant-baseline="middle"
                        font-size={cellSize/4}
                >
                    {modifier}
                </text>
            </g>
        {/if}
        {#if letter}
        <g>
            <rect
                    x={x*cellSize + 2}
                    y={y*cellSize + 2}
                    width={cellSize - 4}
                    height={cellSize - 4}
                    fill="#fff"
                    stroke="rgba(0,0,0,0.5)"
                    rx={cellSize/6}
                    ry={cellSize/6}
                    class="tile"
            />
            <text
                    x={x*cellSize + cellSize/2}
                    y={y*cellSize + cellSize/2}
                    text-anchor="middle"
                    dominant-baseline="central"
                    font-size={cellSize/3}
            >
                {letter}
            </text>
            <text
                    x={(x+1)*cellSize - cellSize/5}
                    y={(y+1)*cellSize - cellSize/4}
                    text-anchor="end"
                    dominant-baseline="central"
                    fill="rgba(0,0,0,0.5)"
                    font-size={cellSize/5}
            >
                {letterPoints[letter]}
            </text>
        </g>
        {/if}
    {/each}
</svg>