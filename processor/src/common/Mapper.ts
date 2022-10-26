export class Mapper {
    constructor() {
        // Prevent instantiation
    }

    public static toId(name: string): string {
        return name.toLowerCase().replace(" ", "");
    }

    public static toProtocol(type: string): string {
        return type.replace("/", "-").toLowerCase();
    }

    public static toUrlProtocol(protocol: string): string | undefined {
        switch (protocol) {
            case "http-rest":
                return "http";
        }
    }
}
