import { Chess } from "chess.js";
export type Square = `${"a" | "b" | "c" | "d" | "e" | "f" | "g" | "h"}${
  | "1"
  | "2"
  | "3"
  | "4"
  | "5"
  | "6"
  | "7"
  | "8"}`;
export type SquareState = "W" | "B" | "_";

export function isValidMove(g: Chess, move: string): Boolean {
  try {
    g.move(move);
    g.undo();
    return true;
  } catch {
    return false;
  }
}

export function indexToSquareName(i: number, j: number): Square {
  return `${"abcdefgh"[j]}${"12345678"[i]}` as Square;
}

export function squareNameToIndex(squareName: Square): [number, number] {
  const i = ["a", "b", "c", "d", "e", "f", "g", "h"].indexOf(squareName[0]!);
  const j = ["1", "2", "3", "4", "5", "6", "7", "8"].indexOf(squareName[1]!);

  return [i, j];
}
