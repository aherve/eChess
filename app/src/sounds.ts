import { logger } from "./logger";
import playSound from "play-sound";

const player = playSound();

export async function playNotifySound() {
  logger.info("Playing notify sound");
  await genericPlay(
    `${__dirname}/../sounds/public_sound_standard_GenericNotify.mp3`
  );
}
export async function playMoveSound() {
  logger.info("Playing move sound");
  await genericPlay(`${__dirname}/../sounds/public_sound_standard_Move.mp3`);
}
export async function playCaptureSound() {
  logger.info("Playing capture sound");
  await genericPlay(`${__dirname}/../sounds/public_sound_standard_Capture.mp3`);
}

export async function playDingSound() {
  logger.info("Playing draw sound");
  await genericPlay(`${__dirname}/../sounds/ding.mp3`);
}

async function genericPlay(path: string): Promise<void> {
  return new Promise((resolve, reject) => {
    player.play(path, (err) => {
      if (err) {
        logger.error(`Could not play sound: ${err}`, err);
        reject(err);
      }
      resolve();
    });
  });
}
