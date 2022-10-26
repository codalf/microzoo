import {StackPort, StackService} from "./StackServiceTransformer";
import spawn, {NoOpCallbacks, ChildProcessResult} from "../common/spawn";
import {ChildProcess} from "child_process";
const treeKill = require("tree-kill");

type PortInfo = {
    service: StackService;
    port: StackPort
};

class PortForwardCallbacks extends NoOpCallbacks {
    constructor(public onExit: (result: ChildProcessResult) => void = () => {}) {
        super();
    }

    handleStderr(data: Buffer): void {
        console.log(data.toString());
        if (data.toString().indexOf("error forwarding port") !== -1) {
            this.onExit({stdout: "", stderr: data.toString(), exitCode: -1, error: new Error("Error forwarding port")});
        }
    }
}

export class PortForwardManager {
    constructor(private namespace: string) {}

    public start(portInfos: PortInfo[]): void {
        try {
            let childProcesses = portInfos.map(({service, port}) => this.startPortForward(service, port, result => {
              if (result.error && childProcesses.length) {
                  console.log("* Port forwarding has stopped, restart to resume");
                  const processes = [...childProcesses];
                  childProcesses = [];
                  this.stopPortForwards(processes.filter(process => process));
              }
            }));
            console.log("> Started port forwarding, press Ctrl+C to terminate");
        }
        catch (_) {
            console.log("* Could not start port forwarding, maybe still active")
        }
    }

    private startPortForward(service: StackService, port: StackPort, onExit: (result: ChildProcessResult) => void): ChildProcess | undefined {
        return spawn(`kubectl port-forward service/${service.id} ${port.targetPort}:${port.sourcePort} --namespace=${this.namespace}`, true, new PortForwardCallbacks(onExit));
    }

    private stopPortForwards(processes: ChildProcess[] | undefined): void {
        processes?.forEach(process => {
            treeKill(process.pid);
        })
    }
}
