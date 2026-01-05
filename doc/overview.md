# Project Overview: Microzoo

Microzoo is a tool designed to process PlantUML specifications (scenarios) and transform them into deployable microservice environments (stacks) on targets like Docker Compose or Kubernetes.

## Folder Structure and Overview

### `processor/` 
The core component of the project, written in TypeScript. It handles parsing, transformation, validation, and deployment logic.

*   **`processor/src/index.ts`**: Entry point. Defines the CLI using `commander`. Registers deployers (`DockerComposeDeployer`, `KubernetesDeployer`) and executes commands like `compile`, `deploy`, `test`, and `drop`.
*   **`processor/src/command/`**: Contains command implementations.
    *   `compile.ts`: Logic for compiling a scenario.
    *   `deploy.ts`: Logic for deploying a system.
    *   `test.ts`: Logic for running tests on a deployed system.
    *   `drop.ts`: Logic for removing a deployed system.
*   **`processor/src/deployment/`**: Target-specific deployment logic.
    *   `MicrozooDeployer.ts`: Defines the `MicrozooDeployer` interface and `DeployerFactory`.
    *   `DockerComposeDeployer.ts`: Implementation for Docker Compose targets.
    *   `KubernetesDeployer.ts`: Implementation for Kubernetes targets.
    *   `StackServiceTransformer.ts`: Transforms system models into stack-specific service configurations.
    *   `PortForwardManager.ts`: Manages networking/access to services (likely for K8s).
*   **`processor/src/transformation/`**: 
    *   `PumlTransformer.ts`: Converts PlantUML models into an internal `MicrozooSystem` model.
*   **`processor/src/manifest/`**:
    *   `ManifestRegistry.ts`: Manages component manifests found in the `components/` directory.
*   **`processor/src/model/`**: Internal data structures.
    *   `PumlSystem.ts`: Model representing the parsed PlantUML.
    *   `MicrozooSystem.ts`: Refined model used for deployment.
    *   `Types.ts`: Common type definitions.
*   **`processor/src/validation/`**:
    *   `PumlValidator.ts`: Validates raw PlantUML input.
    *   `MicrozooValidator.ts`: Validates the transformed system model.
*   **`processor/src/config/`**: Configuration management.
    *   `globalConfig.ts`: Application-wide settings.
    *   `Constants.ts`: Fixed string and numeric constants.
*   **`processor/src/common/`**: Utilities.
    *   `spawn.ts`: Helper for executing shell commands.
    *   `Mapper.ts`: Object mapping utility.
    *   `StringUtil.ts`: String manipulation helpers.
    *   `delay.ts`: Promise-based sleep function.

### `components/` 
Reusable building blocks for the generated systems. Each component has a `microzoo.yml` manifest.

*   **`components/database/`**: Pre-defined database containers (e.g., `mysql`, `mariadb`, `mongodb`, `postgresql`).
*   **`components/service/`**: Microservice templates in various languages (`go-service`, `spring-boot-service`, `quarkus-service`).

### `scenarios/` 
Input files for Microzoo. PlantUML (`.puml`) files describing the desired microservice architecture (e.g., `minimal.puml`, `basic.puml`, `complex.puml`).

### `stacks/` 
Target output directory. Contains generated configuration files (Docker Compose files, Kubernetes manifests) for specific scenarios.

### `doc/` 
Project documentation.

## Dependencies

### Internal Dependencies
- `index.ts` depends on `command/*`, `deployment/*`, and `config/*`.
- `command/*` implementations use `transformation/*`, `validation/*`, `manifest/*`, and `deployment/*` via the `DeployerFactory`.
- `deployment/*` implementations depend on `model/*` and `common/spawn.ts` to execute CLI tools (`docker-compose`, `kubectl`).
- `transformation/PumlTransformer.ts` depends on `model/*`.

### External Dependencies
- **Commander**: CLI framework used in `index.ts`.
- **PlantUML**: Not as a library, but the format is the primary input (parsed by `PumlTransformer`).
- **Docker / Docker Compose**: External dependency for running systems locally.
- **Kubernetes (kubectl / helm)**: External dependency for cluster deployments.
- **File System (fs)**: Heavily used for reading manifests and writing stack files.

## External Service Integrations

- **Databases**: Supported via `components/database/`. Integrated by providing environment variables and network links in the generated stacks.
- **Message Queues**: Not explicitly seen in the high-level folder structure, but could be added as a component.
- **Identity Providers**: Not explicitly seen, though the `complex.puml` scenario might include one. 
