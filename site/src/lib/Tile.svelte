<script lang="ts">
    type Props = {
        x?: number;
        y?: number;
        cellSize: number;
        letter: string;
        points?: number;
        selected?: boolean;
    }

    let { x=0, y=0, cellSize, letter, points, selected }: Props = $props();
</script>

<style>
    svg {
        display: block;
        margin: 0;
        padding: 0;
        filter: drop-shadow( 1px 2px 3px rgba(0, 0, 0, .25));
    }
</style>

<svg
        x={x}
        y={y}
        width={cellSize}
        height={cellSize}
        viewBox={`0 0 ${cellSize} ${cellSize}`}
>
        <rect
                x={2}
                y={2}
                width={cellSize - 4}
                height={cellSize - 4}
                fill="#fff"
                stroke={selected ? "rgba(0,0,255,1)" : "rgba(0,0,0,0.5)"}
                rx={cellSize/6}
                ry={cellSize/6}
                class="tile"
        />
        <text
                x={cellSize/2}
                y={cellSize/2}
                text-anchor="middle"
                dominant-baseline="central"
                font-size={cellSize/2.5}
        >
            {letter}
        </text>
        {#if points}
            <text
                    x={cellSize - cellSize/5}
                    y={cellSize - cellSize/4}
                    text-anchor="end"
                    dominant-baseline="central"
                    fill="rgba(0,0,0,0.75)"
                    font-size={cellSize/5}
            >
                {points}
            </text>
        {/if}
</svg>