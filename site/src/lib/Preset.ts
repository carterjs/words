export type Preset = {
    key: string;
    name: string;
    description: string;
    config: {
        rackSize: number;
        letterDistribution: Record<string, number>;
        letterPoints: Record<string, number>;
        modifiers: Record<string, {
            value: string;
            grids: {
                x: number;
                y: number;
                width: number;
                height: number;
            }[];
            bothDiagonals: {
                startAt: number;
                skipCount: number;
                matchCount: number;
            }[]
        }>[];
    }
}