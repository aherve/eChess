import { it, describe } from "vitest";
import { Gui } from "./gui";

describe("gui integration test", () => {
  it("renders the board with the correct pieces", async () => {
    const gui = new Gui();
    gui.setBoardStatus(true);
    gui.updateBoard([
      //
      ["W", "W", "W", "W", "W", "W", "W", "W"],
      ["W", "W", "W", "W", "W", "W", "W", "W"],
      ["_", "_", "_", "_", "_", "_", "_", "_"],
      ["_", "_", "_", "_", "_", "_", "_", "_"],
      ["_", "_", "_", "_", "_", "_", "_", "_"],
      ["_", "_", "_", "_", "_", "_", "_", "_"],
      ["B", "B", "B", "B", "B", "B", "B", "B"],
      ["B", "B", "B", "B", "B", "B", "B", "B"],
      /*
       *["B", "B", "B", "B", "B", "B", "B", "B"],
       *["B", "B", "B", "B", "B", "B", "B", "B"],
       *["B", "B", "B", "B", "B", "B", "B", "B"],
       *["B", "B", "B", "B", "B", "B", "B", "B"],
       *["B", "B", "B", "B", "B", "B", "B", "B"],
       *["B", "B", "B", "B", "B", "B", "B", "B"],
       *["B", "B", "B", "B", "B", "B", "B", "B"],
       *["B", "B", "B", "B", "B", "B", "B", "B"],
       */
    ]);

    await new Promise((r) => setTimeout(r, 2000));
  });
});
