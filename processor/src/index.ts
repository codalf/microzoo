import {DeployerFactory} from "./deployment/MicrozooDeployer";
import {Command, Option} from 'commander';
import {DockerComposeDeployer} from "./deployment/DockerComposeDeployer";
import {KubernetesDeployer} from "./deployment/KubernetesDeployer";
import compile from "./command/compile";
import deploy from "./command/deploy";
import test from "./command/test";
import drop from "./command/drop";

type Options = {[key: string]: string};

function getSourceFolder(program: Command, options: Options): string {
    return options.sourceFolder || program.opts().sourceFolder;
}

function getTarget(program: Command, options: Options): string {
  return options.target || program.opts().target
}

function buildCompileCommand(program: Command) {
  return new Command('compile')
    .arguments('<source>')
    .description('compiles a puml specification')
    .action((source: string, options) => {
        compile(source, getSourceFolder(program, options), getTarget(program, options))
          .catch(reason => console.log(reason));
    });
}

function buildDeployCommand(program: Command) {
    return new Command('deploy')
      .arguments('<source>')
      .description('compiles, deploys and runs a puml specification')
      .action((source: string, options) => {
          deploy(source, getSourceFolder(program, options), getTarget(program, options))
            .catch(reason => console.log(reason));
      });
}

function buildTestCommand(program: Command) {
    return new Command('test')
      .arguments('<source>')
      .description('compiles, deploys, runs and tests a puml specification')
      .action((source: string, options) => {
          test(source, getSourceFolder(program, options), getTarget(program, options))
            .catch(reason => console.log(reason));
      });
}
function buildDropCommand(program: Command) {
    return new Command('drop')
      .arguments('<source>')
      .description('drops a deployed system')
      .action((source: string, options) => {
          drop(source, getSourceFolder(program, options), getTarget(program, options))
            .catch(reason => console.log(reason));
      });
}

function start(argv: string[]) {
    const program = new Command()
      .version('0.9.0', '-v, --version', 'output the current version')
      .option('-s, --source-folder <folder>', 'set the source folder', '../scenarios')
      .addOption(new Option('-t, --target <type>', 'set the target system')
        .choices(["docker-compose", "kubernetes"]).default('docker-compose'));
    
    program.addCommand(buildCompileCommand(program))
      .addCommand(buildDeployCommand(program))
      .addCommand(buildTestCommand(program))
      .addCommand(buildDropCommand(program));
    program.parse(argv);
}

try {
    DeployerFactory.register("docker-compose", DockerComposeDeployer);
    DeployerFactory.register("kubernetes", KubernetesDeployer);
    start(process.argv);
}
catch(error) {
    console.log(error);
}
