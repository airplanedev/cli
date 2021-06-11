// This file includes a shim that will execute your task code.
import task from "../{{.ImportPath}}";

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
    await task(JSON.parse(process.argv[2]));
  } catch (err) {
    console.error(err);
    console.log(
      "airplane_output:error " + JSON.stringify({ "error": String(err) }),
    );
    process.exit(1);
  }
}

main();
