const ENABLE_LOGGING = process.env.ENABLE_LOGGING === "true" || process.env.ENABLE_LOGGING === "1";
export function log(type = "info", ...args) {
    if (!ENABLE_LOGGING)
        return;
    switch (type) {
        case "info":
            console.info(...args);
            break;
        case "error":
            console.error(...args);
            break;
        default:
            console.log(...args);
    }
}
export function parseSchemaPermissions(permissionsString) {
    const permissions = {};
    if (!permissionsString) {
        return permissions;
    }
    const permissionPairs = permissionsString.split(",");
    for (const pair of permissionPairs) {
        const [schema, value] = pair.split(":");
        if (schema && value) {
            permissions[schema.trim()] = value.trim() === "true";
        }
    }
    return permissions;
}
export function parseMySQLConnectionString(connectionString) {
    const config = {};
    let cleanedString = connectionString.trim().replace(/^mysql\s+/, '');
    const tokens = [];
    let currentToken = '';
    let inQuotes = false;
    let quoteChar = null;
    for (let i = 0; i < cleanedString.length; i++) {
        const char = cleanedString[i];
        if ((char === '"' || char === "'") && (!inQuotes || char === quoteChar)) {
            inQuotes = !inQuotes;
            quoteChar = inQuotes ? char : null;
        }
        else if (char === ' ' && !inQuotes) {
            if (currentToken) {
                tokens.push(currentToken);
                currentToken = '';
            }
        }
        else {
            currentToken += char;
        }
    }
    if (currentToken) {
        tokens.push(currentToken);
    }
    for (let i = 0; i < tokens.length; i++) {
        const token = tokens[i];
        if (token.startsWith('-') && !token.startsWith('--')) {
            const flag = token[1];
            let value = token.substring(2);
            if (!value && i + 1 < tokens.length && !tokens[i + 1].startsWith('-')) {
                value = tokens[i + 1];
                i++;
            }
            switch (flag) {
                case 'h':
                    config.host = value;
                    break;
                case 'P': {
                    const port = parseInt(value, 10);
                    if (Number.isNaN(port) || !Number.isFinite(port) || port < 1 || port > 65535) {
                        throw new Error(`Invalid port: ${value}`);
                    }
                    config.port = port;
                    break;
                }
                case 'u':
                    config.user = value;
                    break;
                case 'p':
                    config.password = value;
                    break;
                case 'S':
                    config.socketPath = value;
                    break;
            }
        }
        else if (token.startsWith('--')) {
            const [flag, ...valueParts] = token.substring(2).split('=');
            let value = valueParts.join('=');
            if (!value && i + 1 < tokens.length && !tokens[i + 1].startsWith('-')) {
                value = tokens[i + 1];
                i++;
            }
            switch (flag) {
                case 'host':
                    config.host = value;
                    break;
                case 'port': {
                    const port = parseInt(value, 10);
                    if (Number.isNaN(port) || !Number.isFinite(port) || port < 1 || port > 65535) {
                        throw new Error(`Invalid port: ${value}`);
                    }
                    config.port = port;
                    break;
                }
                case 'user':
                    config.user = value;
                    break;
                case 'password':
                    config.password = value;
                    break;
                case 'socket':
                    config.socketPath = value;
                    break;
            }
        }
        else if (!token.startsWith('-')) {
            config.database = token;
        }
    }
    return config;
}
