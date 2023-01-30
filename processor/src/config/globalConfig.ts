let globalConfig = {
    "container-cli": "docker",
    "compose-cli": "docker compose"
};

export function setGlobalConfig(config: typeof globalConfig): void {
    globalConfig = config;
}

export function getGlobalConfig(): typeof globalConfig {
    return globalConfig;
}