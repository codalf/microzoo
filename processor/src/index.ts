import {DeployerFactory} from "./deployment/MicrozooDeployer";
import {Command, Option, OptionValues} from 'commander';
import {DockerComposeDeployer} from "./deployment/DockerComposeDeployer";
import {KubernetesDeployer} from "./deployment/KubernetesDeployer";
import {setGlobalConfig} from "./config/globalConfig";
import * as fs from "fs";
import compile from "./command/compile";
import deploy from "./command/deploy";
import test from "./command/test";
import drop from "./command/drop";

type OptionGetter = (options: OptionValues, key: string) => string;

function buildCompileCommand(getOption: OptionGetter) {
  return new Command('compile')
    .arguments('<source>')
    .description('compiles a puml specification')
    .action((source: string, options) => {
        compile(source, getOption(options, "sourceFolder"), getOption(options, "target"))
          .catch(reason => console.log(reason));
    });
}

function buildDeployCommand(getOption: OptionGetter) {
    return new Command('deploy')
      .arguments('<source>')
      .description('compiles, deploys and runs a puml specification')
      .action((source: string, options) => {
          deploy(source, getOption(options, "sourceFolder"), getOption(options, "target"))
            .catch(reason => console.log(reason));
      });
}

function buildTestCommand(getOption: OptionGetter) {
    return new Command('test')
      .arguments('<source>')
      .description('compiles, deploys, runs and tests a puml specification')
      .action((source: string, options) => {
          test(source, getOption(options, "sourceFolder"), getOption(options, "target"))
            .catch(reason => console.log(reason));
      });
}
function buildDropCommand(getOption: OptionGetter) {
    return new Command('drop')
      .arguments('<source>')
      .description('drops a deployed system')
      .action((source: string, options) => {
          drop(source, getOption(options, "sourceFolder"), getOption(options, "target"))
            .catch(reason => console.log(reason));
      });
}

function start(argv: string[]) {
    const program = new Command()
      .version('0.9.0', '-v, --version', 'output the current version')
      .option('-s, --source-folder <folder>', 'set the source folder', '../scenarios')
      .addOption(new Option('-t, --target <type>', 'set the target system')
        .choices(["docker-compose", "kubernetes"]).default('docker-compose'))
      .option('-c, --config-file <file>', 'set the config file', './config.json');

    const getOption = (options: OptionValues, key: string): string => {
        return options[key] || program.opts()[key];
    }  
      
    program.addCommand(buildCompileCommand(getOption))
      .addCommand(buildDeployCommand(getOption))
      .addCommand(buildTestCommand(getOption))
      .addCommand(buildDropCommand(getOption));

    program.hook('preAction', () => {
      const configFile = fs.readFileSync(program.opts().configFile).toString();
      setGlobalConfig(JSON.parse(configFile));
    });

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
