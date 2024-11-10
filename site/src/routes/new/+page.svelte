<script lang="ts">
    import { PUBLIC_API_URL } from '$env/static/public';

    import {afterNavigate, goto, pushState, replaceState} from "$app/navigation";
    import Header from "$lib/Header.svelte";
    import { page } from '$app/stores'
    import {onMount} from "svelte";

    type Preset = {
        id: string;
        name: string;
        rackSize: number;
        letterDistribution: Record<string, number>;
        letterPoints: Record<string, number>;
    };

    let presets = $state<Preset[]>([]);
    let preset = $state<Preset>();

    const defaultOverrides: {
        rackSize?: number;
        letterDistribution: Record<string, number>;
        letterPoints: Record<string, number>;
    } = {
        letterDistribution: {},
        letterPoints: {}
    };

    let overrides = $state(defaultOverrides);

    let totalCount = $derived(Object.entries(preset?.letterDistribution || {}).reduce((acc, [letter, count]) => {
        return acc + (overrides.letterDistribution[letter] || count);
    }, 0));

    // TODO: account for complexity
    let estimatedPlayTimeMinutes = $derived(totalCount / 2 * Math.sqrt((overrides.rackSize || preset?.rackSize || 7)/7))

    onMount(loadPreset);
    afterNavigate(loadPreset);

    async function loadPreset() {
        const response = await fetch(`${PUBLIC_API_URL}/api/v1/presets`);
        presets = await response.json().then(data => data.presets);
        preset = presets.find(preset => preset.id === $page.url.searchParams.get("preset")) || presets[0];

        if (!preset) {
            return
        }

        let presetId = $page.url.searchParams.get("preset") || "standard";

        for (let preset of presets) {
            if (preset.id === presetId) {
                preset.rackSize = preset.rackSize;
                preset.letterDistribution = preset.letterDistribution;
                preset.letterPoints = preset.letterPoints;
                break;
            }
        }

        overrides = defaultOverrides;

        let override = $page.url.searchParams.get("rackSize");
        if (override) {
            overrides.rackSize = Number(override);
        }

        for (let [letter] of Object.entries(preset.letterDistribution)) {
            let override = $page.url.searchParams.get(`${letter.toLowerCase()}Count`);
            if (override) {
                overrides.letterDistribution[letter] = Number(override);
            }
        }

        for (let [letter] of Object.entries(preset.letterPoints)) {
            let override = $page.url.searchParams.get(`${letter.toLowerCase()}Points`);
            if (override) {
                overrides.letterPoints[letter] = Number(override);
            }
        }
    }

    function displayDuration(minutes: number): string {
        let rounded = Math.ceil(minutes / 15) * 15;
        if (rounded < 60) {
            return `${rounded} minutes`;
        } else {
            let hours = Math.floor(rounded / 60);
            let remainderMinutes = rounded % 60;
            return `${hours} hour${hours == 1 ? "" : "s"}${remainderMinutes ? ` and ${remainderMinutes} minutes` : ""}`;
        }
    }

    async function createGame(e: SubmitEvent) {
        e.preventDefault();

        const url = PUBLIC_API_URL + "/api/v1/games";

        console.log(JSON.stringify({
            preset: preset?.id || "standard",
            overrides
        }));

        const resp = await fetch(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                preset: preset?.id || "standard",
                overrides
            })
        });

        if (!resp.ok) {
            console.error(resp.statusText);
            return;
        }

        const { id } = await resp.json();

        // redirect to play interface
        const playUrl = new URL(location.href);
        playUrl.pathname = "/play";
        playUrl.search = "";
        playUrl.searchParams.set("game", id);
        await goto(playUrl);
    }

</script>
<style>
    label {
        display: block;
    }
    input {
        margin-top: 0.5rem;
        margin-bottom: 1rem;
        display: block;
        box-sizing: border-box;
        width: 100%;
        padding: 0.5rem 1rem;
        border: 1px solid rgba(0, 0, 0, 0.25);
        font: inherit;
    }
    table {
        margin-top: 1rem;
        margin-bottom: 2rem;
        width: 100%;
        border-collapse: collapse;
    }
    th {
        padding: 0.5rem;
    }
    td {
        padding: 0;
        min-width: 2rem;
        border: 1px solid rgba(0, 0, 0, 0.25);
        text-align: center;
        background-color: rgba(255, 255, 255, 0.25);
    }
    td input {
        border: none;
        text-align: center;
        margin: 0;
    }
</style>
<Header />
<main class="container">
    {#if preset}
    <form onsubmit={createGame}>
        <h2>Preset</h2>
        <!--TODO: style as big blocks-->
        <ul>
            {#each presets as { id, name }}
                <li>
                    <a href="?preset={id}">{name}</a>
                </li>
            {/each}
        </ul>
        <h2>Configuration</h2>
        <label>
            Rack size
            <input type="number" min="1" value={overrides.rackSize || preset?.rackSize} onchange={(e) => {
                if (!preset) return

                overrides.rackSize = Number(e.target?.value);
                // set in url
                const url = new URL(location.href);
                url.searchParams.set("rackSize", e.target?.value);
                replaceState(url, {});
        }} />
        </label>
        <table>
            <thead>
            <tr>
                <th></th>
                <th>Count</th>
                <th>Points</th>
            </tr>
            </thead>
            <tbody>
            {#each Object.entries(preset.letterDistribution) as [letter, count]}
                <tr>
                    <td>{letter == "_" ? "" : letter}</td>
                    <td>
                        <input type="number" min="0" value={overrides.letterDistribution[letter] || count} onchange={(e) => {
                            overrides.letterDistribution[letter] = Number(e.target?.value);
                            // set in url
                            const url = new URL(location.href);
                            url.searchParams.set(`${letter.toLowerCase()}Count`, overrides.letterDistribution[letter].toString());
                            replaceState(url, {});
                        }} />
                    </td>
                    <td>
                        <input type="number" min="0" value={overrides.letterPoints[letter] || preset.letterPoints[letter]} onchange={(e) => {
                            overrides.letterPoints[letter] = Number(e.target?.value);
                            // set in url
                            const url = new URL(location.href);
                            url.searchParams.set(`${letter.toLowerCase()}Points`, overrides.letterPoints[letter].toString());
                            replaceState(url, {});
                        }} />
                    </td>
                </tr>
            {/each}
            </tbody>
        </table>
        <p>Letter count: {totalCount}</p>
        <p>Estimated play time: {displayDuration(estimatedPlayTimeMinutes)}</p>
        <button class="button">Start game</button>
    </form>
    {/if}
</main>

