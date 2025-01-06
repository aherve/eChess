import { z } from "zod";
import { Game, GameEventSchema, GameSchema } from "./types";
import { logger } from "./logger";

const lichessToken =
  process.env.LICHESS_API_TOKEN ??
  require("../secret.json")["LICHESS_API_TOKEN"];

export async function claimVictory(gameId: string) {
  const claimed = await lichessFetch(
    `board/game/${gameId}/claim-victory`,
    {},
    "POST"
  );
  return claimed.ok;
}

export async function resignGame({ gameId }: { gameId: string }) {
  const resigned = await lichessFetch(
    `board/game/${gameId}/resign`,
    {},
    "POST"
  );
  return resigned.ok;
}
export async function abortGame({ gameId }: { gameId: string }) {
  const aborted = await lichessFetch(`board/game/${gameId}/abort`, {}, "POST");
  return aborted.ok;
}

export async function drawGame({ gameId }: { gameId: string }) {
  const drawed = await lichessFetch(
    `board/game/${gameId}/draw/yes`,
    {},
    "POST"
  );
  return drawed.ok;
}

export async function createSeek({
  time,
  increment,
}: {
  time: number;
  increment: number;
}) {
  logger.info("seeking game", { time, increment });
  const controller = new AbortController();
  const seek = await lichessFetch(
    "board/seek",
    {
      rated: "true",
      time: `${time}`,
      increment: `${increment}`,
      variant: "standard",
      ratingRange: "",
    },
    "POST",
    controller.signal
  );
  if (!seek.ok) {
    logger.error("could not seek game", await seek.text());
    throw new Error("Error while creating seek");
  }
  return controller;
}

export async function playMove(gameId: string, move: string) {
  const played = await lichessFetch(
    `board/game/${gameId}/move/${move}`,
    {},
    "POST"
  );
  return played.ok;
}

export async function* streamGame(gameId: string) {
  const streamRequest = await lichessFetch(`board/game/stream/${gameId}`);
  if (!streamRequest.ok) {
    logger.error(await streamRequest.text());
    throw new Error("Error while opening stream");
  }
  logger.info("lichess stream opened", gameId);

  const reader = streamRequest.body?.getReader();
  if (!reader) {
    throw new Error("No reader found in stream response");
  }
  logger.info("fetching first chunk");
  let readResult = await reader.read();
  while (!readResult.done) {
    const jsonArray = Buffer.from(readResult.value.buffer)
      .toString()
      .split("\n");
    for (const json of jsonArray) {
      if (!json) continue;
      const parsed = GameEventSchema.safeParse(JSON.parse(json));
      if (!parsed.success) {
        logger.error("Could not parse event from lichess", {
          json,
          error: parsed.error,
        });
        throw new Error("Could not parse event from lichess");
      }
      yield parsed.data;
    }

    readResult = await reader.read();
  }
  logger.info("stream ended");
}

export async function findAndWatch(): Promise<Game | null> {
  const resp = await lichessFetch("account/playing", { nb: "1" });
  if (!resp.ok) {
    logger.error(await resp.text());
    throw new Error("Error while fetching playing games");
  }
  const rawJson = await resp.json();
  const data = z
    .object({ nowPlaying: z.array(GameSchema) })
    .parse(await rawJson);
  if (!data.nowPlaying.length) {
    return null;
  }

  logger.info("full game data from lichess", rawJson);
  logger.info("parsed game data ", data.nowPlaying[0]);

  return data.nowPlaying[0];
}

function lichessFetch(
  path: string,
  params?: Record<string, string>,
  method: "GET" | "POST" = "GET",
  signal?: AbortSignal
) {
  const url =
    method === "GET"
      ? `https://lichess.org/api/${path}?${new URLSearchParams(params ?? {})}`
      : `https://lichess.org/api/${path}`;

  const hasBody =
    method === "POST" && !!params && Object.keys(params).length > 0;

  return fetch(url, {
    signal,
    keepalive: true,
    method: method ?? "GET",
    headers: {
      Authorization: `Bearer ${lichessToken}`,
      ...(hasBody
        ? { "Content-Type": "application/x-www-form-urlencoded" }
        : {}),
    },
    ...(hasBody ? { body: new URLSearchParams(params) } : {}),
  });
}
