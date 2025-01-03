import contrib from "blessed-contrib";
import blessed from "blessed";
import { SquareState } from "./utils";
import { asciiPlug, asciiUsb } from "./ascii-art";
import { abortGame, createSeek, drawGame, resignGame } from "./lichess";
import { logger } from "./logger";
import type { Game, GameStateEvent } from "./types";
import { playCaptureSound, playMoveSound, playNotifySound } from "./sounds";
import { Chess } from "chess.js";

type ClockAnchor = {
  wtime: number;
  btime: number;
  setAt: number;
};

export class Gui {
  private hasBoard = false;
  private gameId: string | null = null;
  private screen: blessed.Widgets.Screen;
  private board: Array<Array<SquareState>> = [];
  private grid: contrib.grid;
  private color: "black" | "white" | null = null;
  private isMyTurn = false;
  private clockAnchor: ClockAnchor | null = null;
  private opponentName: string | null = null;
  private seekAbortController: AbortController | null = null;

  constructor() {
    this.screen = blessed.screen({
      smartCSR: true,
      autoPadding: true,
      terminal: "xterm-basic",
      fullUnicode: true,
    });
    this.screen.key(["escape", "q", "C-c"], () => {
      return process.exit(0);
    });
    this.grid = new contrib.grid({ rows: 12, cols: 12, screen: this.screen });
    this.autoRefresh();
  }

  public terminateGame() {
    this.gameId = null;
    this.opponentName = null;
    this.abortSeek();
    playNotifySound();
  }

  public setBoardStatus(hasBoard: boolean) {
    this.hasBoard = hasBoard;
  }

  public startGame(game: Game) {
    playNotifySound();
    this.abortSeek();
    this.color = game.color;
    this.gameId = game.fullId;
    this.opponentName = `${game.opponent.username} (${game.opponent.rating})`;
  }

  public updateFromLichess(event: GameStateEvent) {
    this.playGameSound(event);
    const { btime, wtime, moves } = event;
    this.clockAnchor = {
      setAt: Date.now(),
      wtime: wtime,
      btime: btime,
    };
    this.isMyTurn =
      (this.color === "white" && moves.length % 2 === 0) ||
      (this.color === "black" && moves.length % 2 === 1);
  }

  public updateBoard(board: Array<Array<SquareState>>) {
    this.board = board;
  }

  private async playGameSound(event: GameStateEvent) {
    switch (event.status) {
      case "mate":
      case "resign":
      case "draw":
      case "timeout":
      case "aborted":
      case "nostart":
      case "outoftime":
      case "cheat":
      case "unknownfinish":
      case "stalemate":
      case "created":
        return playNotifySound();
      case "started": {
        const lichessMoves = event.moves;
        if (!lichessMoves.length) {
          return playNotifySound();
        }
        const g = new Chess();
        for (const move of lichessMoves.slice(0, -1)) {
          g.move(move);
        }
        const lastMove = g.move(lichessMoves[lichessMoves.length - 1]);
        if (lastMove.captured) {
          return playCaptureSound();
        } else {
          return playMoveSound();
        }
      }
    }
  }

  private async autoRefresh() {
    this.render();
    await new Promise((resolve) => setTimeout(resolve, 100));
    this.autoRefresh();
  }

  private render() {
    if (!this.hasBoard) {
      return this.renderNoBoard();
    }
    if (!this.gameId) {
      return this.renderNoGame();
    }
    this.renderGame();
  }

  private renderGame() {
    if (!this.color) {
      return;
    }
    const opponentColor = this.color === "white" ? "black" : "white";
    this.grid.set(0, 0, 6, 10, blessed.box, {
      align: "center",
      content:
        `${opponentColor}: ${this.opponentName}` +
        "\n" +
        this.getPrettyTime({
          forColor: opponentColor,
          isPlaying: !this.isMyTurn,
        }) +
        (this.isMyTurn ? "  " : " \u{1F7E2}"),
    });
    this.grid.set(6, 0, 6, 10, blessed.box, {
      align: "center",
      content:
        `You play ${this.color}` +
        "\n" +
        this.getPrettyTime({ forColor: this.color, isPlaying: this.isMyTurn }) +
        (this.isMyTurn ? " \u{1F7E2}" : "  "),
    });
    const resign = this.grid.set(0, 10, 4, 2, blessed.button, {
      top: "center",
      align: "center",
      left: "center",
      content: "Resign 🏁",
      mouse: true,
    });
    resign.on(
      "press",
      () => this.gameId && resignGame({ gameId: this.gameId })
    );

    const abort = this.grid.set(4, 10, 4, 2, blessed.button, {
      top: "center",
      align: "center",
      left: "center",
      content: "Abort ❌",
      mouse: true,
    });
    abort.on("press", () => this.gameId && abortGame({ gameId: this.gameId }));

    const draw = this.grid.set(8, 10, 4, 2, blessed.button, {
      top: "center",
      align: "center",
      left: "center",
      content: "Draw ½",
      mouse: true,
    });
    draw.on("press", () => this.gameId && drawGame({ gameId: this.gameId }));

    this.screen.render();
  }

  private renderNoGame() {
    const box = this.grid.set(0, 0, 12, 12, blessed.box, {
      align: "center",
      border: {
        type: "line",
      },
      style: {
        fg: "white",
        border: {
          fg: "#f0f0f0",
        },
      },
    });

    if (this.seekAbortController) {
      box.content = "Seeking a game...";
      const abort = this.grid.set(4, 4, 4, 4, blessed.button, {
        align: "center",
        top: "center",
        left: "center",
        content: "Abort",
        mouse: true,
      });
      abort.on("press", () => this.abortSeek());
    } else if (isUnpoweredBoard(this.board)) {
      box.content = "Please connect the board to a power source." + asciiPlug;
    } else if (!isStartingPosition(this.board)) {
      box.content = "Please place the pieces in the starting position.";

      blessed.box({
        align: "center",
        parent: box,
        top: "center",
        left: "center",
        width: "50%",
        height: "50%",
        content: buildAsciiBoard(this.board),
      });
    } else {
      box.content = "Ready for a new game";
      const fifteenTen = this.grid.set(4, 1, 5, 5, blessed.button, {
        top: "center",
        align: "center",
        content: "Create 15 | 10 game",
        left: "center",
        mouse: true,
      });
      const tenZero = this.grid.set(4, 6, 5, 5, blessed.button, {
        top: "center",
        align: "center",
        content: "Create 10 | 0 game",
        left: "center",
        mouse: true,
      });

      fifteenTen.on("press", () => this.seekGame({ time: 15, increment: 10 }));
      tenZero.on("press", () => this.seekGame({ time: 10, increment: 0 }));
    }

    this.screen.render();
  }

  private abortSeek() {
    if (this.seekAbortController) {
      this.seekAbortController.abort();
      this.seekAbortController = null;
    }
  }

  private seekGame({ time, increment }: { time: number; increment: number }) {
    createSeek({ time, increment })
      .then((ctrl) => {
        this.seekAbortController = ctrl;
      })
      .catch((e) => {
        logger.error(e);
      });
  }

  private renderNoBoard() {
    this.grid.set(0, 0, 12, 12, blessed.box, {
      top: "center",
      align: "center",
      left: "center",
      width: "95%",
      height: "95%",
      content: `Connecting to the board...` + asciiUsb,
      tags: true,
      border: {
        type: "line",
      },
      style: {
        fg: "white",
        border: {
          fg: "#f0f0f0",
        },
      },
    });

    this.screen.render();
  }

  private getPrettyTime({
    forColor,
    isPlaying,
  }: {
    forColor: "white" | "black";
    isPlaying: boolean;
  }): string {
    if (this.clockAnchor === null) {
      return "";
    }
    const baseTime =
      forColor === "white" ? this.clockAnchor.wtime : this.clockAnchor.btime;

    if (isPlaying) {
      const elapsed = Date.now() - this.clockAnchor.setAt;
      const time = baseTime - elapsed;
      return prettyTimer(time);
    } else {
      return prettyTimer(baseTime);
    }
  }
}

function buildAsciiBoard(board: Array<Array<SquareState>>): string {
  let res = "";
  for (let i = 7; i >= 0; i--) {
    res += "|" + board[i].join("|") + "|\n";
  }
  return res;
}

function isUnpoweredBoard(board: Array<Array<SquareState>>): boolean {
  return board.every((row) => row.every((cell) => cell === "B"));
}

function isStartingPosition(board: Array<Array<SquareState>>): boolean {
  // 2 rows of white pieces
  for (const i of [0, 1]) {
    if (board[i].some((cell) => cell !== "W")) {
      return false;
    }
  }
  for (const i of [2, 5]) {
    if (board[i].some((cell) => cell !== "_")) {
      return false;
    }
  }
  for (const i of [6, 7]) {
    if (board[i].some((cell) => cell !== "B")) {
      return false;
    }
  }
  return true;
}

function prettyTimer(msTime: number): string {
  const time = msTime / 1000;

  const hours = Math.floor(time / 3600);
  const minutes = Math.floor((time % 3600) / 60);
  const seconds = Math.floor(time % 60);

  if (hours > 0) {
    return `${hours.toString().padStart(2, "0")}:${minutes
      .toString()
      .padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
  }
  return `${minutes.toString().padStart(2, "0")}:${seconds
    .toString()
    .padStart(2, "0")}`;
}
