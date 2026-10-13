// Runs a Postman collection with Newman and exits as soon as the run is done (Newman alone can linger on Windows).
import newman from 'newman';

const [collection, environment, workingDir] = process.argv.slice(2);

newman.run(
  { collection, environment, workingDir, reporters: 'cli', color: 'off', disableUnicode: true },
  (error, summary) => {
    if (error) {
      console.error(error);
      process.exit(2);
    }
    process.exit(summary.run.error || summary.run.failures.length > 0 ? 1 : 0);
  },
);
