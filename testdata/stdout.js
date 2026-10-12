
const fs = require("fs");
const { argv } = require("process");

if (argv.length < 3) {
  console.error("Usage: node stdout.js <wasm_file.wasm>");
  process.exit(1);
}

const wasmFile = argv[2];
const bytes = fs.readFileSync(wasmFile);


(async () => {
  const obj = await WebAssembly.instantiate(new Uint8Array(bytes));
  const result = obj.instance.exports.Run();
  console.log(result);
})();
