import os from "os";
import winston from "winston";

const today = new Date().toISOString().slice(0, 10);

export const logger = winston.createLogger({
  level: "info",
  format: winston.format.simple(),
  transports: [
    /*
     *new winston.transports.Console({
     *  handleExceptions: true,
     *  handleRejections: true,
     *}),
     */
    new winston.transports.File({
      filename: `${os.homedir()}/echess/${today}.log`,
      handleExceptions: true,
      handleRejections: true,
    }),
  ],
});
