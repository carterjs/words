import { PUBLIC_API_URL } from '$env/static/public';

type board = {
    cells: Cell[]
}

export type Cell = {
    x: number;
    y: number;
    modifier?: "TW" | "DL" | "TL" | "DW";
    letter?: string;
}

export type Player = {
    id: string;
    name: string;
    score: number;
}

export type Placement = {
    x: number;
    y: number;
    direction: "HORIZONTAL" | "VERTICAL";
    word: string;
    points: number;
    indirectWords: { x: number; y: number; direction: string; word: string }[] | null;
}

export type Challenge = {
    challengerId: string;
    moverId: string;
    votesInvalid: number;
    votesValid: number;
    votesNeeded: number;
    eligibleVoters: number;
}

export type LastWord = {
    playerId: string;
    x: number;
    y: number;
    direction: "HORIZONTAL" | "VERTICAL";
    word: string;
}

export type Announcement = {
    text: string;
    // board cell to jump to when the announcement is tapped
    at?: { x: number; y: number };
}

export  class GameController {
    id = $state<string>("");

    loaded = $state<boolean>(false);

    playerId = $state<string | null>(null);

    #rack = $state<string[]>([]);

    get rack() {
        return this.#rack;
    }

    set rack(value: string[]) {
        // sort by points
        value.sort((a, b) => {
            return this.letterPoints[b] - this.letterPoints[a];
        });

        this.#rack = value;
        this.sortedRack = value;
    }

    updateSortedRack() {
        this.sortedRack = this.sortedRack.sort((a, b) => {
            // stable sort where letters not in the rack are sorted to the end
            // adding 1 makes sure all letters in the word are sorted in placement order (including index 0)
            // || 999 replaces all -1 (now 0) values with 999 so they go to the end
            return (this.input.indexOf(a)+1 || 999) - (this.input.indexOf(b)+1 || 999);
        });
    }

    sortedRack = $state<string[]>([]);

    #input = $state<string>("");

    get input() {
        return this.#input;
    }

    set input(value: string) {
        this.#input = value;
        this.updateSortedRack();
    }

    started = $state<boolean>(false);

    finished = $state<boolean>(false);

    winnerIds = $state<string[]>([]);

    currentPlayerId = $state<string>("");

    lettersRemaining = $state<number>(0);

    players = $state<Player[]>([]);

    challenge = $state<Challenge | null>(null);

    // the id of the player who played the still-challengeable last word
    challengeableMoverId = $state<string | null>(null);

    // the most recently played word, highlighted on the board until the next move
    lastWord = $state<LastWord | null>(null);

    announcement = $state<Announcement | null>(null);

    #announcementTimer: ReturnType<typeof setTimeout> | undefined;

    announce(text: string, at?: { x: number; y: number }) {
        this.announcement = { text, at };
        clearTimeout(this.#announcementTimer);
        this.#announcementTimer = setTimeout(() => {
            this.announcement = null;
        }, 8000);
    }

    myVote = $state<boolean>(false);

    placements = $state<Placement[]>([]);

    selectedPlacement = $state<number>(0);

    error = $state<string>("");

    board = $state<board>({ cells: [] });

    letterPoints = $state<{ [letter: string]: number }>({})

    get myTurn() {
        return this.playerId !== null && this.playerId === this.currentPlayerId;
    }

    getPlayerById(id: string) {
        return this.players.find(p => p.id === id);
    }

    playerName(id: string) {
        return this.getPlayerById(id)?.name ?? "someone";
    }

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
            finished: boolean;
            currentPlayerId: string;
            lettersRemaining: number;
            winnerIds?: string[];
            challenge?: Challenge;
            challengeableMoverId?: string;
            players: Player[];
            playerId: string | null;
            rack?: string[];
            letterPoints: { [letter: string]: number };
        } = await resp.json();

        this.started = data.started;
        this.finished = data.finished;
        this.currentPlayerId = data.currentPlayerId;
        this.lettersRemaining = data.lettersRemaining;
        this.winnerIds = data.winnerIds || [];
        this.challenge = data.challenge || null;
        this.challengeableMoverId = data.challengeableMoverId || null;
        this.playerId = data.playerId || null;
        this.letterPoints = data.letterPoints;
        this.rack = data.rack || [];
        this.players = data.players || [];

        await this.reloadBoard();

        this.loaded = true;
    }

    async reloadBoard() {
        let resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}/board`, {
            credentials: "include"
        });
        if (!resp.ok) {
            throw new Error(`Failed to fetch board for game ${this.id}`);
        }

        const { cells } = await resp.json();
        this.board = { cells: cells || [] };
    }

    async loadBoard(minX: number, minY: number, maxX: number, maxY: number) {
        let resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}/board?minX=${minX}&minY=${minY}&maxX=${maxX}&maxY=${maxY}`, {
            credentials: "include"
        });
        if (!resp.ok) {
            throw new Error(`Failed to fetch grid for game ${this.id}`);
        }

        let { cells } = await resp.json();

        // lazy-loaded ranges can overlap what's already loaded
        const known = new Set(this.board.cells.map((cell: Cell) => `${cell.x},${cell.y}`));
        const fresh = (cells || []).filter((cell: Cell) => !known.has(`${cell.x},${cell.y}`));

        this.board.cells = this.board.cells.concat(fresh);
    }

    #events: EventSource | null = null;

    streamUpdates() {
        // the server decides at connection time whether we get our private
        // events (like rack updates), so reconnect after identity changes
        this.#events?.close();

        // server sent events
        let events = new EventSource(`${PUBLIC_API_URL}/api/v1/games/${this.id}/events`, {
            withCredentials: true
        });
        this.#events = events;

        // catch up on anything that happened while not connected - a reopen
        // after joining, a dropped connection, or a server restart
        events.onopen = async () => {
            if (!this.loaded) return;
            await this.refreshMeta();
            await this.reloadBoard();
        };

        events.addEventListener("GAME_STARTED", (e) => {
            this.started = true;

            const { letters } = JSON.parse(e.data);
            if (letters?.length) {
                this.rack = letters;
            }

            // the game snapshot changed under us; pick up turn order
            this.refreshMeta();
        })

        events.addEventListener("PLAYER_JOINED", (e) => {
            const { playerId, playerName } = JSON.parse(e.data);
            if (!this.getPlayerById(playerId)) {
                this.players.push({ id: playerId, name: playerName, score: 0 });
            }
        })

        events.addEventListener("WORD_PLAYED", (e) => {
            const played: { playerId: string, nextPlayerId: string, points: number } & Placement = JSON.parse(e.data);

            this.mergeWord(played);
            this.currentPlayerId = played.nextPlayerId;
            this.challengeableMoverId = played.playerId;
            this.myVote = false;

            // the board changed under any placement someone was previewing
            if (played.playerId !== this.playerId) {
                this.clearPlacements();
            }

            this.lastWord = {
                playerId: played.playerId,
                x: played.x,
                y: played.y,
                direction: played.direction,
                word: played.word,
            };

            if (played.playerId !== this.playerId) {
                const middle = Math.floor(played.word.length / 2);
                this.announce(
                    `${this.playerName(played.playerId)} played ${played.word} for ${played.points} point${played.points === 1 ? "" : "s"}`,
                    {
                        x: played.direction === "HORIZONTAL" ? played.x + middle : played.x,
                        y: played.direction === "VERTICAL" ? played.y + middle : played.y,
                    },
                );
            }

            const player = this.getPlayerById(played.playerId);
            if (player) {
                player.score += played.points;
            }

            this.refreshMeta();
        })

        events.addEventListener("RACK_UPDATED", (e) => {
            const { letters } = JSON.parse(e.data);
            this.rack = letters || [];
        })

        events.addEventListener("TURN_PASSED", (e) => {
            const { playerId, nextPlayerId } = JSON.parse(e.data);
            this.currentPlayerId = nextPlayerId;
            this.challengeableMoverId = null;
            this.lastWord = null;
            if (playerId !== this.playerId) {
                this.announce(`${this.playerName(playerId)} skipped their turn`);
            }
            this.refreshMeta();
        })

        events.addEventListener("LETTERS_EXCHANGED", (e) => {
            const { playerId, count, nextPlayerId } = JSON.parse(e.data);
            this.currentPlayerId = nextPlayerId;
            this.challengeableMoverId = null;
            this.lastWord = null;
            if (playerId !== this.playerId) {
                this.announce(`${this.playerName(playerId)} swapped ${count} letter${count === 1 ? "" : "s"}`);
            }
            this.refreshMeta();
        })

        events.addEventListener("CHALLENGE_STARTED", (e) => {
            this.challenge = JSON.parse(e.data);
        })

        events.addEventListener("CHALLENGE_VOTE_CAST", (e) => {
            const { votesInvalid, votesValid, votesNeeded } = JSON.parse(e.data);
            if (this.challenge) {
                this.challenge = { ...this.challenge, votesInvalid, votesValid, votesNeeded };
            }
        })

        events.addEventListener("CHALLENGE_RESOLVED", async (e) => {
            const { upheld, rescindedWord } = JSON.parse(e.data);
            this.challenge = null;
            this.challengeableMoverId = null;
            this.lastWord = null;
            this.myVote = false;

            if (upheld) {
                this.announce(`${rescindedWord ? `"${rescindedWord}"` : "The word"} was voted invalid and removed`);
                // the word came off the board; scores changed too
                await this.reloadBoard();
            } else {
                this.announce("The challenge failed - the word stands");
            }
            await this.refreshMeta();
        })

        events.addEventListener("GAME_ENDED", (e) => {
            const { winnerIds, scores } = JSON.parse(e.data);
            this.finished = true;
            this.winnerIds = winnerIds || [];
            for (const player of this.players) {
                if (scores && player.id in scores) {
                    player.score = scores[player.id];
                }
            }
        })
    }

    // refreshMeta re-syncs scores, pool size, and turn from the server. It's
    // cheap and keeps locally-tracked numbers honest.
    async refreshMeta() {
        let resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}`, {
            credentials: "include"
        });
        if (!resp.ok) {
            return;
        }

        const data = await resp.json();
        this.started = data.started;
        this.players = data.players || [];
        this.currentPlayerId = data.currentPlayerId;
        this.lettersRemaining = data.lettersRemaining;
        this.finished = data.finished;
        this.winnerIds = data.winnerIds || [];
        this.challenge = data.challenge || null;
        this.challengeableMoverId = data.challengeableMoverId || null;

        // safety net in case a private rack event was missed
        if (data.rack && [...data.rack].sort().join() !== [...this.rack].sort().join()) {
            this.rack = data.rack;
        }
    }

    mergeWord(placement: Placement) {
        const cells = [...this.board.cells];
        const letters = [...placement.word];

        for (let i = 0; i < letters.length; i++) {
            const x = placement.direction === "HORIZONTAL" ? placement.x + i : placement.x;
            const y = placement.direction === "VERTICAL" ? placement.y + i : placement.y;

            const existing = cells.find(cell => cell.x === x && cell.y === y);
            if (existing) {
                existing.letter = letters[i];
            } else {
                cells.push({ x, y, letter: letters[i] });
            }
        }

        this.board = { cells };
    }

    // tapLetter adds an unused rack letter to the word being built, or
    // removes a used one from it
    tapLetter(letter: string, used: boolean) {
        if (used) {
            const index = this.input.lastIndexOf(letter);
            if (index !== -1) {
                this.input = this.input.slice(0, index) + this.input.slice(index + 1);
            }
        } else {
            this.input += letter;
        }
    }

    setRackOrder(letters: string[]) {
        this.sortedRack = letters;
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
            this.error = await this.errorMessage(resp, "Couldn't join the game.");
            return;
        }

        let { playerId, players } = await resp.json();

        this.playerId = playerId;
        this.players = players || [];

        // the event stream was opened before we had an identity; reconnect so
        // the server includes our private events
        this.streamUpdates();
    }

    async start() {
        let resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}`, {
            credentials: "include",
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
        await this.refreshMeta();
    }

    async findPlacements(x: number, y: number) {
        this.error = "";
        this.placements = [];
        this.selectedPlacement = 0;

        let resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}/board/placements?x=${x}&y=${y}&word=${encodeURIComponent(this.input)}`, {
            credentials: "include"
        });

        if (!resp.ok) {
            this.error = await this.errorMessage(resp, "That word can't go there.");
            return;
        }

        this.placements = await resp.json();
    }

    clearPlacements() {
        this.placements = [];
        this.selectedPlacement = 0;
        this.error = "";
    }

    async playSelectedPlacement() {
        const placement = this.placements[this.selectedPlacement];
        if (!placement) return;

        let resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}/board`, {
            credentials: "include",
            method: "PATCH",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                operation: "ADD_WORD",
                payload: {
                    x: placement.x,
                    y: placement.y,
                    direction: placement.direction,
                    word: placement.word,
                }
            })
        });

        if (!resp.ok) {
            this.error = await this.errorMessage(resp, "Couldn't play that word.");
            return;
        }

        // board, rack, scores, and turn all update via events
        this.input = "";
        this.clearPlacements();
    }

    async updateGame(operation: string, payload?: object) {
        this.error = "";

        let resp = await fetch(`${PUBLIC_API_URL}/api/v1/games/${this.id}`, {
            credentials: "include",
            method: "PATCH",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({ operation, payload })
        });

        if (!resp.ok) {
            this.error = await this.errorMessage(resp, "That didn't work.");
            return null;
        }

        return await resp.json();
    }

    async pass() {
        await this.updateGame("PASS_TURN");
    }

    async exchange(letters: string[]) {
        const result = await this.updateGame("EXCHANGE_LETTERS", { letters });
        if (result) {
            this.input = "";
        }
    }

    async challengeWord() {
        const result = await this.updateGame("CHALLENGE_WORD");
        if (result) {
            this.myVote = true;
        }
    }

    async castVote(vote: "VALID" | "INVALID") {
        const result = await this.updateGame("CAST_VOTE", { vote });
        if (result) {
            this.myVote = true;
        }
    }

    async errorMessage(resp: Response, fallback: string) {
        try {
            const { error } = await resp.json();
            return error || fallback;
        } catch {
            return fallback;
        }
    }
}
