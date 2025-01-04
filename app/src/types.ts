import { z } from "zod";

export const opponentGoneEventSchema = z.object({
  type: z.literal("opponentGone"),
  gone: z.boolean(),
  claimWinInSeconds: z.number(),
});

export const ChatLineEventShema = z.object({
  type: z.literal("chatLine"),
  room: z.enum(["player", "spectator"]),
  text: z.string(),
  username: z.string(),
});
export type ChatLineEvent = z.infer<typeof ChatLineEventShema>;

export const GameStateEventSchema = z.object({
  type: z.literal("gameState"),
  moves: z.string().transform((x) =>
    x
      .split(" ")
      .map((s) => s.trim())
      .filter((x) => x.length > 0)
  ),
  wtime: z.number(),
  btime: z.number(),
  status: z.enum([
    "aborted",
    "cheat",
    "created",
    "draw",
    "mate",
    "nostart",
    "outoftime",
    "resign",
    "stalemate",
    "started",
    "timeout",
    "unknownfinish",
    "variantend",
  ]),
  winner: z.enum(["white", "black"]).optional(),
});
export type GameStateEvent = z.infer<typeof GameStateEventSchema>;

export const GameFullEventSchema = z.object({
  type: z.literal("gameFull"),
  state: GameStateEventSchema,
  color: z.enum(["black", "white"]).optional(),
  //white: z.object({
  //id: z.string(),
  //username: z.string(),
  //}),
  //black: z.object({
  //id: z.string(),
  //username: z.string(),
  //}),
  //id: z.string(),
  //variant: z.object({
  //  key: z.string(),
  //  name: z.string(),
  //}),
  //speed: z.string(),
  //perf: z.string(),
  //rated: z.boolean(),
  //createdAt: z.number(),
  //lastMoveAt: z.number(),
  //status: z.object({
  //  id: z.number(),
  //  name: z.string(),
  //}),
  //turns: z.number(),
  //fen: z.string(),
  //lastMove: z.string(),
  //source: z.string(),
  //clock: z.object({
  //  initial: z.number(),
  //  increment: z.number(),
  //}),
  //players: z.object({
  //  white: z.object({
  //    id: z.string(),
  //    rating: z.number(),
  //  }),
  //  black: z.object({
  //    id: z.string(),
  //    rating: z.number(),
  //  }),
  //}),
  //spectator: z.boolean(),
  //winner: z.enum(["white", "black"]).optional(),
});

export const GameEventSchema = z.discriminatedUnion("type", [
  GameStateEventSchema,
  GameFullEventSchema,
  ChatLineEventShema,
  opponentGoneEventSchema,
]);
export type GameEvent = z.infer<typeof GameEventSchema>;

export const GameSchema = z.object({
  fullId: z.string(),
  gameId: z.string(),
  fen: z.string(),
  color: z.enum(["black", "white"]),
  lastMove: z.string(),
  source: z.string(),
  status: z.object({
    id: z.number(),
    name: z.string(),
  }),
  variant: z.object({
    key: z.string(),
    name: z.string(),
  }),
  speed: z.string(),
  perf: z.string(),
  rated: z.boolean(),
  hasMoved: z.boolean(),
  opponent: z.object({
    id: z.string().nullable(),
    username: z.string(),
    rating: z.number().default(0),
    //ai: z.number(),
  }),
  isMyTurn: z.boolean(),
});
export type Game = z.infer<typeof GameSchema>;
