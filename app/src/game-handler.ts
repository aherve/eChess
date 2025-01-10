import { SerialPort } from "serialport";
import { Chess, type Color, type Square } from "chess.js";
import { claimVictory, playMove } from "./lichess";
import {
  type SquareState,
  indexToSquareName,
  isValidMove,
  squareNameToIndex,
} from "./utils";
import { logger } from "./logger";

const LEDdebounceDelay = 100; // don't make LED state change like crazy in case a measurement gets unstable
const playDelay = 250; // wait before sending a detected move to lichess. Allows to slide a piece through multiple squares without playing them

export class GameHandler {
  private lichessMoves: Array<string>;
  private arduinoReady = false;
  private litSquares: Set<Square>;
  private arduinoBoard?: Array<Array<SquareState>>;
  private lichessGameId?: string;
  private myColor?: Color;
  private lastPlayed?: string;
  private debounceState = {
    lastSentAt: 0,
    isQueued: false,
  };
  private candidateMove?: {
    move: string;
    playAt: number;
  };

  constructor(private port: SerialPort) {
    this.arduinoReady = false;
    this.lichessMoves = [];
    this.litSquares = new Set<Square>();
    this.sendLEDCommand();
  }

  public async reset(gameId: string, color: Color) {
    this.lastPlayed = undefined;
    this.myColor = color;
    this.lichessGameId = gameId;
    this.lichessMoves = [];
    await this.startSequence();
    return this.reconcile();
  }

  public async terminateGame() {
    this.lichessGameId = undefined;
    await this.endSequence();
  }

  public async claimVictory() {
    if (!this.lichessGameId) {
      logger.info("No game id found, can't claim victory");
      return;
    }
    logger.info("Opponent is gone for good. Claiming victory");
    return claimVictory(this.lichessGameId);
  }

  public async updateArduinoBoard(newBoard: Array<Array<SquareState>>) {
    this.arduinoReady = true;
    this.arduinoBoard = newBoard;
    this.reconcile();
  }

  public async updateLichessMoves(moves: Array<string>) {
    logger.info("updating lichess moves", moves);
    this.lichessMoves = moves.filter((m) => m.length);
    this.reconcile();
  }

  private async startSequence() {
    const period = 40;
    let prev: Square | null = null;
    let curr: Square | null = null;
    for (let i = 2; i < 6; i++) {
      for (let j = 0; j < 8; j++) {
        prev = curr;
        curr = indexToSquareName(i, j);
        this.litSquares = new Set([prev, curr].filter((x): x is Square => !!x));
        this.sendLEDCommandNow();
        await new Promise((resolve) => setTimeout(resolve, period));
      }
      i++;
      for (let j = 7; j >= 0; j--) {
        prev = curr;
        curr = indexToSquareName(i, j);
        this.litSquares = new Set([prev, curr].filter((x): x is Square => !!x));
        this.sendLEDCommandNow();
        await new Promise((resolve) => setTimeout(resolve, period));
      }
    }
    this.litSquares = new Set();
    this.sendLEDCommandNow();
  }

  private async endSequence() {
    const period = 300;
    for (let i = 0; i < 3; i++) {
      this.litSquares = new Set(["d4", "d5", "e4", "e5"]);
      this.sendLEDCommandNow();
      await new Promise((resolve) => setTimeout(resolve, period));
      this.litSquares = new Set([]);
      this.sendLEDCommandNow();
      await new Promise((resolve) => setTimeout(resolve, period));
    }
  }

  private async sendLEDCommand() {
    const now = Date.now();
    if (now - this.debounceState.lastSentAt > LEDdebounceDelay) {
      this.debounceState.lastSentAt = now;
      this.debounceState.isQueued = false;
      return this.sendLEDCommandNow();
    }
    if (this.debounceState.isQueued) {
      return;
    }

    this.debounceState.isQueued = true;
    setTimeout(() => {
      this.sendLEDCommand();
    }, now - this.debounceState.lastSentAt + LEDdebounceDelay + 5);
  }

  private async sendLEDCommandNow() {
    const commandBuffer = Buffer.alloc(this.litSquares.size + 2);
    commandBuffer[0] = 254;
    commandBuffer[commandBuffer.length - 1] = 255;
    Array.from(this.litSquares).forEach((square, pos) => {
      const [i, j] = squareNameToIndex(square);
      commandBuffer[pos + 1] = (j << 4) + i;
    });
    while (!this.arduinoReady) {
      await new Promise((resolve) => setTimeout(resolve, 200));
    }
    return await new Promise<void>((resolve, reject) => {
      this.port.write(commandBuffer, (err) => {
        if (err) {
          logger.error("Error while writing to serial port", err);
          return reject(err);
        }
        return resolve();
      });
    });
  }

  private async reconcile() {
    // Build board from lichess moves
    const g = new Chess();
    for (const move of this.lichessMoves) {
      g.move(move);
    }
    const lichessState = new Map<Square, SquareState>();
    const occupiedSquares = g.board().flat();
    for (const occupiedSquare of occupiedSquares) {
      if (!occupiedSquare) {
        continue;
      }
      lichessState.set(
        occupiedSquare.square as Square,
        occupiedSquare.color === "w" ? "W" : "B"
      );
    }

    // find diff between arduinoBoard and lichessState
    const lit = new Set<Square>();
    const possibleSources = new Set<Square>();
    const possibleDestinations = new Set<Square>();
    if (this.lichessGameId) {
      for (let i = 0; i < 8; i++) {
        for (let j = 0; j < 8; j++) {
          const square = indexToSquareName(i, j);
          const lichessSquareState = lichessState.get(square) ?? "_";
          const arduinoSquareState = this.arduinoBoard?.[i]?.[j];
          if (lichessSquareState !== arduinoSquareState) {
            lit.add(square);

            if (lichessSquareState && arduinoSquareState === "_") {
              possibleSources.add(square);
            } else {
              possibleDestinations.add(square);
            }
          }
        }
      }
      // send LED state to the board
      this.litSquares = lit;
    } else {
      this.litSquares = new Set();
    }
    this.sendLEDCommand();

    // if arduino describes a unique, valid move, and it's our turn, then we play the move
    const myTurn = g.turn() === this.myColor;
    if (!myTurn) {
      logger.info("waiting for opponent to play");
      this.candidateMove = undefined;
      return;
    }

    if (possibleSources.size === 0 && possibleDestinations.size === 0) {
      logger.info("position matches lichess state");
      this.candidateMove = undefined;
      return;
    }

    if (possibleSources.size !== 1 || possibleDestinations.size !== 1) {
      this.candidateMove = undefined;
      return;
    }

    const candidateMove =
      Array.from(possibleSources)[0]! + Array.from(possibleDestinations)[0];

    if (!candidateMove) {
      logger.info(" Wait, what ?", { possibleSources, possibleDestinations });
      this.candidateMove = undefined;
      return;
    }

    if (!isValidMove(g, candidateMove)) {
      logger.info("invalid move", candidateMove);
      this.candidateMove = undefined;
      return;
    }

    // Make sure we don't spam lichess api
    if (this.lastPlayed === candidateMove) {
      logger.info("already played this move", candidateMove);
      this.candidateMove = undefined;
      return;
    }

    // Schedule a move play
    this.playWithDelay(candidateMove);
  }

  // Allow the user to swipe a piece through multiple valid squares withouth playing them as soon as they are detected
  private async playWithDelay(newMove?: string) {
    const now = Date.now();
    if (!this.lichessGameId) {
      logger.error("No game id found");
      return;
    }

    // record new move for later
    if (newMove) {
      this.candidateMove = {
        move: newMove,
        playAt: Date.now() + playDelay,
      };
    } else {
      // is there a move to play ?
      if (this.candidateMove && this.candidateMove.playAt < now) {
        const toPlay = this.candidateMove.move;
        await playMove(this.lichessGameId, toPlay);
        this.lastPlayed = toPlay;
        this.candidateMove = undefined;
        return;
      }
    }

    if (!this.candidateMove) {
      return;
    }

    setTimeout(() => this.playWithDelay(), this.candidateMove.playAt - now + 1);
  }
}
