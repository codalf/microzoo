import {Readable, Writable} from "stream";
import childProcess, {ChildProcess, ChildProcessByStdio} from "child_process";

export interface ChildProcessResult {
    stdout: string;
    stderr: string;
    exitCode?: number;
    error?: Error;
}

export interface ChildProcessCallbacks {
    handleStdout(data: Buffer): void;
    handleStderr(data: Buffer): void;
    handleClose(exitCode: number): void;
    onExit(result: ChildProcessResult): void;
}

export class NoOpCallbacks implements  ChildProcessCallbacks {
    handleStdout(data: Buffer): void {}
    handleStderr(data: Buffer): void {}
    handleClose(exitCode: number): void {}
    onExit(result: ChildProcessResult): void {}
}

class StreamCollector implements ChildProcessCallbacks {
    private stdout = "";
    private stderr = "";

    constructor(public onExit: (result: ChildProcessResult) => void = () => {}) {}

    handleStdout(data: Buffer): void {
        this.stdout += data;
    }

    handleStderr(data: Buffer): void {
        this.stderr += data;
    }

    handleClose(exitCode: number): void {
        if (exitCode) {
            const error = new Error('Shell command exit with non zero code: ' + exitCode)
            this.onExit({stdout: this.stdout, stderr: this.stderr, exitCode: exitCode, error});
        } else {
            this.onExit({stdout: this.stdout, stderr: this.stderr, exitCode: 0});
        }
    }
}

const defSpawnOptions = { stdio: 'inherit' }

/**
 * @summary Get shell program meta for current platform
 * @private
 * @returns {Object}
 */
function getShell(): {cmd: string, arg: string} {
    if (process.platform === 'win32') {
        return { cmd: 'cmd', arg: '/C' }
    } else {
        return { cmd: 'sh', arg: '-c' }
    }
}

/**
 * Callback is called with the output when the process terminates. Output is
 * available when true is passed as options argument or stdio: null set
 * within given options.
 *
 * @summary Execute shell command forwarding all stdio
 * @param {String|Array} command
 * @param {Object|true} [options] spawn() options or true to set stdio: null
 * @param {Object} [callbacks]
 * @returns {ChildProcess}
 */
function spawn(command: string[] | string, options: {stdio: string | null} | boolean, callbacks: ChildProcessCallbacks = new NoOpCallbacks()): ChildProcessByStdio<Writable, Readable, Readable> | undefined {
    function initEventHandlers(child: ChildProcess): void {
        child.stdout?.on('data', (data: Buffer) => callbacks.handleStdout(data));
        child.stderr?.on('data', (data: Buffer) => callbacks.handleStderr(data));
        child.on('close', (returnCode: number) => callbacks.handleClose(returnCode));
    }

    if (Array.isArray(command)) {
        command = command.join(';');
    }

    if (options === true) {
        options = { stdio: null };
    }
    else {
        options = {...defSpawnOptions, ...options};
    }

    try {
        const shell = getShell();
        const child = childProcess.spawn(shell.cmd, [shell.arg, command], options);
        initEventHandlers(child);
        return child;
    }
    catch (error) {
        callbacks.onExit({stdout: "", stderr: "", error, exitCode: -1});
    }
}

namespace spawn {
    export function promise(command: string | string[], options: {stdio: string | null} | boolean): Promise<ChildProcessResult> {
        return new Promise<ChildProcessResult>((resolve, reject) => {
            spawn(command, options, new StreamCollector(result => {
                if (result.error) {
                    return reject(result);
                }
                else {
                    resolve(result);
                }
            }));
        });
    }
}

export default spawn;
