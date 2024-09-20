const board = document.getElementById("board");
const cellSize = 60;

// websocket connection
let ws;
function connect() {
    ws = new WebSocket("/ws")
    ws.onopen = () => {
        let gameId = window.location.pathname === '/' ? null : window.location.pathname.split('/')[1]
        if (gameId) {
            const playerId = localStorage.getItem(gameId+"/playerId");
            if (!playerId) {

                // set path back to /
                window.history.pushState({}, "", "/");
                return;
            }

            console.log("player", playerId, "rejoining game", gameId);

            ws.send(JSON.stringify({
                type: "rejoin_game",
                payload: {
                    gameId,
                    playerId
                }
            }))
        } else {
            const playSection = document.getElementById("play");
            playSection.hidden = false;
            console.log("hidden board");
            board.hidden = true;
        }
    }

    let game;
    ws.onmessage = (event) => {
       const message = JSON.parse(event.data);

        switch (message.type) {
            case "join_game":
                game = new Game(message.payload.gameId, message.payload.playerId);
                game.players = message.payload.players;
                game.started = message.payload.started;
                game.letterPoints = message.payload.letterPoints;
                break;
            case "create_game":
                game = new Game(message.payload.gameId, message.payload.playerId);
                game.players = message.payload.players;
                game.started = false
                game.letterPoints = message.payload.letterPoints;
                break;
            case "rejoin_game":
                game = new Game(message.payload.gameId, message.payload.playerId);
                game.players = message.payload.players;
                game.started = message.payload.started;
                game.rack = message.payload.rack || [];
                game.letterPoints = message.payload.letterPoints;
                console.log(message.payload);
                game.grid = message.payload.grid;
                break;
            case "start_game":
                game.started = true;
                game.rack = message.payload.rack || [];
                game.grid = message.payload.grid;
                break;
            default:
                console.log("Unknown message", message);
        }
    }
    ws.onclose = () => {
        console.log("disconnected, trying to reconnect in 1 second");
        // try to reconnect
        setTimeout(connect, 1000);
    }
    ws.onerror = (error) => {
        console.log(error);
    }
}
connect()

// game class
class Game {
    _grid = {}
    _localGrid = {}
    _letterPoints = {}
    _selectedOnBoard;
    _selectedInRack;

    _rack = []

    // svg viewBox
    _x = -window.innerWidth/2
    _offsetX = 0
    _y = -window.innerHeight/2
    _offsetY = 0
    _width = window.innerWidth
    _height = window.innerHeight
    _scale = 1

    constructor(gameId, playerId) {
        localStorage.setItem(gameId+"/playerId", playerId);
        window.history.pushState({}, "", `/${gameId}`);

        const gameSection = document.getElementById("game");
        gameSection.hidden = false;
        const playSection = document.getElementById("play");
        playSection.hidden = true;

        this.initializeBoardPanning();
        this.initializeResizing();
    }

    set width(width) {
        this._width = width;
        this.updateViewBox();
    }

    get width() {
        return this._width;
    }

    set height(height) {
        this._height = height;
        this.updateViewBox();
    }

    get height() {
        return this._height;
    }

    set x(x) {
        this._x = x;
        this.updateViewBox();
    }

    get x() {
        return this._x;
    }

    set offsetX(offsetX) {
        this._offsetX = offsetX;
        this.updateViewBox();
    }

    get offsetX() {
        return this._offsetX;
    }

    set y(y) {
        this._y = y;
        this.updateViewBox();
    }

    get y() {
        return this._y;
    }

    set offsetY(offsetY) {
        this._offsetY = offsetY;
        this.updateViewBox();
    }

    get offsetY() {
        return this._offsetY;
    }

    set scale(scale) {
        this._scale = scale;
        this.updateViewBox();
    }

    get scale() {
        return this._scale;
    }

    updateViewBox() {
        board.setAttribute("viewBox", `${(this.x+this.offsetX)*this.scale} ${(this.y+this.offsetY)*this.scale} ${this.width*this.scale} ${this.height*this.scale}`);
    }

    set started(started) {
        const startButton = document.getElementById("start");
        startButton.hidden = started;
    }

    set players(players) {
        const playersList = document.getElementById("players");
        playersList.innerHTML = "";
        players.forEach(player => {
            const playerElement = document.createElement("li");
            playerElement.innerText = player.name;
            playersList.appendChild(playerElement);
        });
    }

    set rack(rack) {
        this._rack = rack;

        const rackList = document.getElementById("rack");
        rackList.innerHTML = "";
        for (let letter of rack) {
            const letterElement = document.createElement("li");
            letterElement.addEventListener("click", () => {
                this.selectedInRack = letter;
            })
            if (letter == "_") {
                letter = " ";
            }
            letterElement.innerText = letter;
            rackList.appendChild(letterElement);
        }
    }

    get rack() {
        return this._rack;
    }

    set grid(grid) {
        let minX = -8;
        let minY = -8;
        let maxX = 8;
        let maxY = 8;

        // bounds
        let computed = {...grid, ...this.localGrid}
        for (let key in computed) {
            let [x,y] = key.split(",");
            x = Number(x)
            y = Number(y)
            // TODO: find existing cells on board and populate those

            if (computed[key] && !this.isModifier(computed[key])) {
                minX = Math.min(minX, x-1);
                minY = Math.min(minY, y-1);
                maxX = Math.max(maxX, x+1);
                maxY = Math.max(maxY, y+1);
            }
        }

        // set values
        for (let key in grid) {
            this._grid[key] = grid[key];
        }

        // now recreate table
        board.innerHTML = "";

        for (let y = minY; y <= maxY; y++) {
            for (let x = minX; x <= maxX; x++) {
                const group = document.createElementNS("http://www.w3.org/2000/svg", "g");
                group.setAttribute("font-size", `${cellSize/2}`);

                let value = computed[`${x},${y}`];

                const cell = document.createElementNS("http://www.w3.org/2000/svg", "rect");
                cell.setAttribute("x", `${x*cellSize+3}`);
                cell.setAttribute("y", `${y*cellSize+3}`);
                cell.setAttribute("width", `${cellSize-6}`);
                cell.setAttribute("height", `${cellSize-6}`);
                cell.setAttribute("fill", "rgba(255,255,255,0.25)");
                cell.setAttribute("rx", "7");
                cell.setAttribute("ry", "7");
                cell.setAttribute("stroke", "#444");

                if (value) {
                    cell.setAttribute("fill", "rgba(255,255,255,0.9)");
                }

                if (this.isModifier(value)) {
                    group.setAttribute("font-size", `${cellSize/3}`);

                    switch (value) {
                        case "DW":
                            cell.setAttribute("fill", "#faf");
                            break;
                        case "DL":
                            cell.setAttribute("fill", "#aaf");
                            break;
                        case "TW":
                            cell.setAttribute("fill", "#faa");
                            break;
                        case "TL":
                            cell.setAttribute("fill", "#ffa");
                            break;
                    }
                }

                group.appendChild(cell);

                const text = document.createElementNS("http://www.w3.org/2000/svg", "text");
                text.setAttribute("x", `${(x+0.5)*cellSize}`);
                text.setAttribute("y", `${(y+0.5)*cellSize}`);
                text.setAttribute("text-anchor", "middle");
                text.setAttribute("dominant-baseline", "central");
                if (x == 0 && y == 0 && !value) {
                    value = "★";
                    cell.setAttribute("fill", "rgba(255,255,255,0.5)");
                    text.setAttribute("fill", "#fff")
                }
                text.appendChild(document.createTextNode(value || ""));
                group.appendChild(text);
                board.appendChild(group);
            }
        }
    }

    set localGrid(overrides) {
        this._localGrid = overrides;
        this.grid = this._grid;
    }

    get localGrid() {
        return this._localGrid;
    }

    isModifier(value) {
        return value === "DW" || value === "DL" || value === "TW" || value === "TL";
    }

    get letterPoints() {
        return this._letterPoints
    }

    set letterPoints(points) {
        this._letterPoints = points;
    }

    set selectedOnBoard(key) {
        if (this.selectedInRack && key) {
            // play the letter from the rack
            console.log("bruh");
            this.localGrid = {...this.localGrid, [key]:this.selectedInRack}
            this.rack = this.rack.filter(l => l !== this.selectedInRack);
            this._selectedInRack = null;
            this._selectedOnBoard = null;
        } else if(this._selectedOnBoard == key) {
            console.log("deselecting");
            this._selectedOnBoard = null
        } else {
            console.log("only board selected");
            this._selectedOnBoard = key
        }
    }
    get selectedOnBoard() {
        return this._selectedOnBoard;
    }

    set selectedInRack(letter) {
        if (this.selectedOnBoard && letter) {
            console.log("both selected");
            this.localGrid = {...this.localGrid, [this.selectedOnBoard]:letter}
            this.rack = this.rack.filter(l => l !== letter);
            this._selectedInRack = null;
            this._selectedOnBoard = null;
        } else if (this._selectedInRack == letter) {
            console.log('deselecting');
            this._selectedInRack = null
        } else {
            console.log('only rack selected');
            this._selectedInRack = letter
        }
    }

    get selectedInRack() {
        return this._selectedInRack
    }

    initializeBoardPanning() {
        let startX;
        let startY;
        board.addEventListener("mousedown", (e) => {
            startX = e.offsetX;
            startY = e.offsetY;
        })
        board.addEventListener("mousemove", (e) => {
            if (!startX || !startY) {
                return;
            }

            this.offsetX = startX - e.offsetX;
            this.offsetY = startY - e.offsetY;
        })
        board.addEventListener("mouseup", (e) => {
            if (Math.abs(this.offsetX) < 3 && Math.abs(this.offsetY) < 3) {
                let clickedX = (e.offsetX + this.x)/cellSize;
                let clickedY = (e.offsetY + this.y)/cellSize;
                let cellClickedX = cellSize * Math.abs(clickedX - Math.floor(clickedX));
                let cellClickedY = cellSize * Math.abs(clickedY - Math.floor(clickedY));

                if (cellClickedX > 3 && cellClickedX < cellSize-3 && cellClickedY > 3 && cellClickedY < cellSize-3) {
                    // this.localGrid = {...this.localGrid, [`${Math.floor(clickedX)},${Math.floor(clickedY)}`]:"F"}
                    this.selectedOnBoard = `${Math.floor(clickedX)},${Math.floor(clickedY)}`;
                }
            }

            this.x = this.x + this.offsetX;
            this.y = this.y + this.offsetY;
            this.offsetX = 0;
            this.offsetY = 0;

            startX = undefined;
            startY = undefined;
        })
    }

    initializeResizing() {
        let timeout;
        board.setAttribute("viewBox", `${this.x} ${this.y} ${this.width} ${this.height}`);
        board.setAttribute("width", `${this.width}`);
        board.setAttribute("height", `${this.height}`);

        window.addEventListener("resize", () => {
            clearTimeout(timeout);
            timeout = setTimeout(() => {
                this.width = window.innerWidth;
                this.height = window.innerHeight;
                board.setAttribute("viewBox", `${this.x} ${this.y} ${this.width} ${this.height}`);
                board.setAttribute("width", `${this.width}`);
                board.setAttribute("height", `${this.height}`);

            }, 10)
        })
    }
}

function joinGame(e) {
    e.preventDefault();

    const formData = new FormData(e.target);
    const gameId = formData.get("gameId");
    const playerName = formData.get("playerName");

    ws.send(JSON.stringify({
        type: "join_game",
        payload: {
            gameId,
            playerName
        }
    }))
}
document.getElementById("join-game").addEventListener("submit", joinGame);

function createGame(e) {
    e.preventDefault();

    const formData = new FormData(e.target);
    const playerName = formData.get("playerName");

    ws.send(JSON.stringify({
        type: "create_game",
        payload: {
            playerName
        }
    }))
}
document.getElementById("create-game").addEventListener("submit", createGame);

function startGame() {
    ws.send(JSON.stringify({
        type: "start_game",
        payload: {}
    }))
}
document.getElementById("start").addEventListener("click", startGame);