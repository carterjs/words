<script lang="ts">
    type Props = {
        x?: number;
        y?: number;
        cellSize: number;
        modifier: "DL" | "TL" | "DW" | "TW";
    }

    let { x=0, y=0, cellSize, modifier }: Props = $props();
</script>

<style>
    svg {
        display: block;
        margin: 0;
        padding: 0;
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
            x={0}
            y={0}
            width={cellSize}
            height={cellSize}
            fill={{
                "DW": "rgba(255,170,255,0.75)",
                "DL": "rgba(170,170,255,0.75)",
                "TW": "rgba(255,170,170,0.75)",
                "TL": "rgba(255,170,170,0.75)",
            }[modifier]}
            stroke="rgba(0,0,0,0.25)"
    />
    <text
            x={cellSize/2}
            y={cellSize/2}
            text-anchor="middle"
            dominant-baseline="middle"
            font-size={cellSize/4}
    >
        {modifier}
    </text>
</svg>