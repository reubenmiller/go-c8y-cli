/**
 * WASI is enabled through asconfig.json in post 0.20.x versions of AssemblyScript.
 * The module, @assemblyscript/wasi-shim is required to enable a WASI environment.
 * 
 * This demo is meant to showcase some abilities of WASI utilized through the AssemblyScript language.
 * It uses the latest version of AssemblyScript and will not work in older (<0.20.x) versions.
**/

import { JSON } from "json-as";

@external("env", "check")
declare function envCheck(i: i32): void // { "log": { "integer"(i) { ... } } }

@external("env", "current_year")
declare function envCurrentYear(): i32 // { "log": { "integer"(i) { ... } } }

@json
class Extension {
  name: string;
  @alias("default")
  default_value: string;
}

// Import module that allows colors in the console as long as supported by WASI.
import { rainbow } from "as-rainbow/assembly";

if (process.argv.length > 1) {
    const name = process.argv[1];
    if (name === "run") {
        run();
    } else if (name === "help") {
        help();
    } else if (name === "options") {
        options();
    }
}

function waitForUserEnter(text: string): void {
    console.log(rainbow.italicMk(text));
    const buf = new ArrayBuffer(1);
    process.stdin.read(buf);
    return;
}

export function help(): void {
    console.log(`
c8y-as [command] [--help]
`);
}

export function options(): string {
    console.log("Calling host function");
    envCheck(1234);

    const currentYear = envCurrentYear();
    console.log(`Current Year: ${currentYear}`);

    // Read from standard input
    const buffer = new ArrayBuffer(1024);
    const total_read = process.stdin.read(buffer);
    console.log(`Read: len=${total_read}, value=${String.UTF8.decode(buffer)}`);

    const extension: Extension = {
        name: "hello",
        default_value: "again",
    };
    console.log(`platform: ${process.platform}`);

    // Read from environment variables
    const keys = process.env.keys();
    for (let i = 0; i < keys.length; ++i) {
        const key = keys[i];
        const value = process.env.get(key);
        console.log(`${key}=${value}`);
    }
    const serialized = JSON.stringify<Extension>(extension);
    console.log(serialized);
    return serialized;
}

export function run(): void {
    // Call console.log which is overridden to use WASI syscalls by @assemblyscript/wasi-shim.
    console.log(rainbow.blue(rainbow.boldMk("Hello from WasmTime WASI through AssemblyScript!")));

    console.log(rainbow.boldMk("Press ctrl+c to exit this demo at any time."));

    // Get the current time and print it.
    console.log("\nWASI can read the system time and display it with the AssemblyScript Date API");
    console.log(rainbow.boldMk("The current time is: ") + rainbow.italicMk(new Date(Date.now()).toString()));

    // Read user input from process.stdin (API provided by AssemblyScript Wasi-Shim and WasmTime WASI API).

    // Create a buffer to hold up to 100 characters of user input.
    const buffer = new ArrayBuffer(100);
    console.log("\nWASI can read user input from stdin. Write any word and return");

    // Read data from stdin and write to buffer.
    process.stdin.read(buffer);

    // Print text and decode the buffer with String.UTF8.decode
    console.log(rainbow.red("You said: " + String.UTF8.decode(buffer)))

    waitForUserEnter("\nPlease press return to continue");

    // Demonstrate the retrieval of cryptographically-safe random numbers through WASI.
    console.log("WASI can fetch cryptographically-safe random numbers from the runtime\nSeries of random numbers:");

    for (let i = 0; i < 5; i++) {
        console.log(Math.random().toString());
    }

    // End
    console.log(rainbow.red("\nThat's all! :D"));
}
