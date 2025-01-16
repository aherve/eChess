#include <SPI.h>

#define boardSize 8
#define ledLatchPin 4
#define hallLatchPin 5

int boardState[boardSize][boardSize];

int const upTres = 570;
int const downTres = 475;
bool ledState[boardSize][boardSize];
int const readPins[boardSize]{ A0, A1, A2, A3, A4, A5, A6, A7 };
byte ledReadCursor;
bool ledStateBuffer[boardSize][boardSize];

void setup() {
  for (int i = 0; i < boardSize; i++) {
    pinMode(readPins[i], INPUT);
    pinMode(ledLatchPin, OUTPUT);
    pinMode(hallLatchPin, OUTPUT);
    digitalWrite(hallLatchPin, LOW);
    digitalWrite(ledLatchPin, LOW);
  }
  SPI.begin();
  resetLEDs();
  Serial.begin(115200);
  //calibrate();
  readBoard(true);
  sendBoard();
}

void loop() {
  if (readBoard(true)) {
    sendBoard();
  }

  readLedState();
  displayLEDs();
}

// expects an array of bytes representing the state of the LEDs
// Command starts with a 254 byte
// each following byte represents a lit square: 4 least significant bits represent the column, 4 most significant bits represent the row
// Command ends with a 255 byte

// Both 254 and 255 are not valid values for a square, so they can be used as signals
void readLedState() {
  if (Serial.available()) {
    byte incomingByte = Serial.read();
    // Reset buffer on start signal
    if (incomingByte == 254) {
      for (byte i = 0; i < boardSize; i++) {
        for (byte j = 0; j < boardSize; j++) {
          ledStateBuffer[i][j] = false;
        }
      }
      return;  // only do one thing at a time
    }
    // Apply state on end signal
    if (incomingByte == 255) {
      for (byte i = 0; i < boardSize; i++) {
        for (byte j = 0; j < boardSize; j++) {
          ledState[i][j] = ledStateBuffer[i][j];
        }
      }
      return;
    }
    // read the byte and push it to the buffer
    byte iByte = incomingByte >> 4;
    byte jByte = incomingByte & 0b00001111;
    if (iByte < boardSize && jByte < boardSize) {
      ledStateBuffer[iByte][jByte] = true;
    }
  }
}

bool readBoard(bool transform) {
  int changed = false;
  byte iByte = 1;
  for (byte i = 0; i < boardSize; i++) {
    SPI.transfer(iByte);
    digitalWrite(hallLatchPin, HIGH);
    digitalWrite(hallLatchPin, LOW);
    for (byte j = 0; j < boardSize; j++) {
      int read = analogRead(readPins[j]);
      int initRead = read;
      if (transform) {
        if (read < downTres) {
          read = -1;
        } else if (read > upTres) {
          read = 1;
        } else {
          read = 0;
        }
      }
      if (read != boardState[i][j]) {
        changed = true;
        boardState[i][j] = read;
      }
    }
    iByte <<= 1;
  }
  return changed;
}

void displayLEDs() {
  byte iByte = 1;
  for (byte i = 0; i < boardSize; i++) {
    bool isLit = false;
    byte jByte = 0;
    for (byte j = 0; j < boardSize; j++) {
      if (ledState[i][j]) {
        isLit = true;
      } else {
        jByte |= 1 << j;  // 1 is off, 0 is on
      }
    }

    SPI.transfer(jByte);
    SPI.transfer(isLit ? iByte : 0);
    SPI.transfer(0);

    digitalWrite(ledLatchPin, HIGH);
    digitalWrite(ledLatchPin, LOW);

    if (isLit) {
      delay(2);  // gotta wait a bit here for the LEDs to be visible
    }

    iByte <<= 1;
  }
  resetLEDs();
}

void resetLEDs() {
  SPI.transfer(255);
  SPI.transfer(0);
  SPI.transfer(0);

  digitalWrite(ledLatchPin, HIGH);
  digitalWrite(ledLatchPin, LOW);
}

void calibrate() {
  int maxBoard[boardSize][boardSize];
  int minBoard[boardSize][boardSize];

  Serial.println("calibrating...");

  int minOfMin = 1024;
  int maxOfMax = 0;

  for (byte i = 0; i < boardSize; i++) {
    for (byte j = 0; j < boardSize; j++) {
      maxBoard[i][j] = 0;
      minBoard[i][j] = 1024;
    }
  }

  int const nbSamples = 100;
  for (byte sample = 0; sample < nbSamples; sample++) {
    displayLEDs();
    readBoard(false);
    for (byte i = 0; i < boardSize; i++) {
      for (byte j = 0; j < boardSize; j++) {
        if (boardState[i][j] < minBoard[i][j]) {
          minBoard[i][j] = boardState[i][j];
        }
        if (boardState[i][j] > maxBoard[i][j]) {
          maxBoard[i][j] = boardState[i][j];
        }
        if (boardState[i][j] < minOfMin) {
          minOfMin = boardState[i][j];
        }
        if (boardState[i][j] > maxOfMax) {
          maxOfMax = boardState[i][j];
        }
      }
    }
  }

  Serial.println("min board");
  for (byte i = 0; i < boardSize; i++) {
    for (byte j = 0; j < boardSize; j++) {
      Serial.print(minBoard[i][j]);
      Serial.print(", ");
    }
    Serial.println();
  }
  Serial.println("max board");
  for (byte i = 0; i < boardSize; i++) {
    for (byte j = 0; j < boardSize; j++) {
      Serial.print(maxBoard[i][j]);
      Serial.print(", ");
    }
    Serial.println();
  }
  Serial.print("max of max: ");
  Serial.println(maxOfMax);
  Serial.print("min of min: ");
  Serial.println(minOfMin);
}

void sendBoard() {
  for (int i = 0; i < boardSize; i++) {
    byte whiteByte = 0;
    byte blackByte = 0;
    for (int j = 0; j < boardSize; j++) {
      if (boardState[i][j] == 1) {
        whiteByte |= 1 << j;
      } else if (boardState[i][j] == -1) {
        blackByte |= 1 << j;
      }
    }
    Serial.write(whiteByte);
    Serial.write(blackByte);
  }
  // End message signal
  // triple 255 would represent at least one row full of both white and plack pieces, which is impossible. This can thus be interpreted as a recognizable signal for the end of the message
  Serial.write(255);
  Serial.write(255);
  Serial.write(255);
}
