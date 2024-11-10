import {onMount} from "svelte";

import { PUBLIC_API_URL } from '$env/static/public';

type player = {
    id: string;
    name: string;
}

type board = {
    cells: Cell[]
}

export type Cell = {
    x: number;
    y: number;
    modifier?: "TW" | "DL" | "TL" | "DW";
    letter?: string;
}

export  class GameController {
    id = $state<string>("");
    loaded = $state<boolean>(false);
    playerId = $state<string | null>(null);
    rack = $state<string[]>([]);
    started = $state<boolean>(false);
    players = $state<{ id: string, name: string }[]>([]);
    messages = $state<string[]>([]);
    board = $state<board>({ cells: [] });

    async loadInitialData() {
        let resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}`, {
            credentials: "include"
        });
        if (!resp.ok) {
            throw new Error(`Failed to fetch game ${this.id}`);
        }

        let data: {
            id: string;
            started: boolean;
            players: { id: string, name: string }[];
            playerId: string | null;
            rack: string[];
        } = await resp.json();

        this.started = data.started;
        this.playerId = data.playerId;
        this.rack = data.rack;
        this.players = data.players;

        resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}/board`, {
            credentials: "include"
        });
        if (!resp.ok) {
            throw new Error(`Failed to fetch board for game ${this.id}`);
        }

        this.board = await resp.json();

        this.loaded = true;
    }

    async loadBoard(minX: number, minY: number, maxX: number, maxY: number) {
        console.log("loading board", minX, minY, maxX, maxY);

        let resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}/board?minX=${minX}&minY=${minY}&maxX=${maxX}&maxY=${maxY}`, {
            credentials: "include"
        });
        if (!resp.ok) {
            throw new Error(`Failed to fetch grid for game ${this.id}`);
        }

        let { cells } = await resp.json();

        this.board.cells = this.board.cells.concat(cells);
    }

    streamUpdates() {
        // server sent events
        let events = new EventSource(`${PUBLIC_API_URL}/api/v1/games/${this.id}/events`, {
            withCredentials: true
        });
        events.addEventListener("MESSAGE", (e) => {
            const { playerId, message } = JSON.parse(e.data);

           this.messages.push(`Player ${playerId} says: ${message}`);
        })
        events.addEventListener("GAME_STARTED", (e) => {
            console.log("received game started event");
            this.started = true;
        })
    }

    getPlayerById(id: string) {
        return this.players.find(p => p.id === id);
    }

    async join(name: string) {
        let resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}`, {
            credentials: "include",
            method: "PATCH",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                operation: "JOIN_GAME",
                payload: {
                    playerName: name,
                }
            })
        });

        if (!resp.ok) {
            console.error("failed to join game", await resp.json());
            throw new Error(`Failed to join game ${this.id}`);
        }

        let { playerId, players } = await resp.json();

        this.playerId = playerId;
        this.players = players;
    }

    async start() {
        let resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}`, {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                operation: "START_GAME"
            })
        });

        if (!resp.ok) {
            console.error("failed to start game", await resp.json());
            throw new Error(`Failed to start game ${this.id}`);
        }

        let { started, rack } = await resp.json();

        this.started = started;
        this.rack = rack;
    }
}