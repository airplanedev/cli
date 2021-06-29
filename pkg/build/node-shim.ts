// This file includes a shim that will execute your task code.
import task from "{{.Entrypoint}}";

async function main() {
  if (process.argv.length !== 3) {
    console.log(
      "airplane_output:error " +
        JSON.stringify({
          "error":
            `Expected to receive a single argument (via {{ "{{JSON}}" }}). Task CLI arguments may be misconfigured.`,
        }),
    );
    process.exit(1);
  }

  try {
    // The task function will either be:
    //  - () => void
    //  - () => Promise<void>
    //  - (params: object) => void
    //  - (params: object) => Promise<void>
    //
    // At build-time, TS will know what the type of the task function is.
    // Therefore, cast as `any` so that TS doesn't throw an error if
    // task does not expect a parameter.
    //
    // We could add a type assertion here to enforce that users don't supply
    // a function that doesn't match one of the signatures above, however
    // the TS error that users would see would not be easy to read, even with
    // TS familiarity.
    await (task as any)(JSON.parse(process.argv[2]));
  } catch (err) {
    console.error(err);
    console.log(
      "airplane_output:error " + JSON.stringify({ "error": String(err) }),
    );
    process.exit(1);
  }
}

main();
